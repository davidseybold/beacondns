package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/oklog/run"

	"github.com/davidseybold/beacondns/internal/api"
	"github.com/davidseybold/beacondns/internal/db/kvstore"
	"github.com/davidseybold/beacondns/internal/db/postgres"
	"github.com/davidseybold/beacondns/internal/dnsstore"
	"github.com/davidseybold/beacondns/internal/log"
	"github.com/davidseybold/beacondns/internal/repository"
	"github.com/davidseybold/beacondns/internal/responsepolicy"
	"github.com/davidseybold/beacondns/internal/worker"
	"github.com/davidseybold/beacondns/internal/zone"
)

type serviceConfig struct {
	Port            int      `env:"BEACON_CONTROLLER_PORT"  envDefault:"8080"`
	DBHost          string   `env:"BEACON_DB_HOST"`
	DBName          string   `env:"BEACON_DB_NAME"          envDefault:"beacon_db"`
	DBUser          string   `env:"BEACON_DB_USER"          envDefault:"beacon_controller"`
	DBPass          string   `env:"BEACON_DB_PASSWORD"`
	DBPort          int      `env:"BEACON_DB_PORT"          envDefault:"5432"`
	ShutdownTimeout int      `env:"BEACON_SHUTDOWN_TIMEOUT" envDefault:"30"`
	EtcdEndpoints   []string `env:"BEACON_ETCD_ENDPOINTS"`
}

func (c *serviceConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}
	if c.DBPort <= 0 || c.DBPort > 65535 {
		return fmt.Errorf("invalid database port number: %d", c.DBPort)
	}

	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("invalid shutdown timeout: %d", c.ShutdownTimeout)
	}
	return nil
}

func main() {
	ctx := context.Background()
	if err := start(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func start(ctx context.Context, w io.Writer) error {
	var err error

	logger := log.NewJSONLogger(slog.LevelInfo, w)

	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	db, err := postgres.NewConnectionPool(ctx, postgres.Config{
		Host:     cfg.DBHost,
		DBName:   cfg.DBName,
		User:     cfg.DBUser,
		Password: cfg.DBPass,
		Port:     cfg.DBPort,
	})
	if err != nil {
		return fmt.Errorf("error creating connection pool: %w", err)
	}
	defer db.Close()

	repoRegistry := repository.NewPostgresRepositoryRegistry(db)
	kvstore, err := kvstore.NewEtcdClient(cfg.EtcdEndpoints, kvstore.Scope{
		Namespace: "beacon",
	})
	if err != nil {
		return fmt.Errorf("error creating etcd client: %w", err)
	}
	defer kvstore.Close()

	dnsStore := dnsstore.New(kvstore)

	zoneService := zone.NewService(repoRegistry)
	zoneEventProcessor := zone.NewEventProcessor(&zone.EventProcessorDeps{
		Repository: repoRegistry,
		DNSStore:   dnsStore,
		Logger:     logger,
	})

	responsePolicyService := responsepolicy.NewService(repoRegistry)
	responsePolicyEventProcessor := responsepolicy.NewEventProcessor(&responsepolicy.EventProcessorDeps{
		Repository: repoRegistry,
		Logger:     logger,
	})

	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	worker := worker.New(
		repoRegistry,
		logger,
		[]worker.EventProcessor{zoneEventProcessor, responsePolicyEventProcessor},
	)

	handler, err := api.NewHTTPHandler(logger, zoneService, responsePolicyService)
	if err != nil {
		return fmt.Errorf("error creating HTTP handler: %w", err)
	}

	var g run.Group
	{
		httpServer := &http.Server{
			//TODO: I just picked a number, I don't know if this is a good value
			ReadHeaderTimeout: time.Second * 10, //nolint:mnd,nolintlint
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           handler,
		}
		g.Add(
			func() error {
				if err = httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}
				return nil
			},
			func(_ error) {
				shutdownCtx, cancel := context.WithTimeout(
					context.Background(),
					time.Duration(cfg.ShutdownTimeout)*time.Second,
				)
				defer cancel()
				if err = httpServer.Shutdown(shutdownCtx); err != nil {
					fmt.Fprintf(w, "error shutting down HTTP server: %s\n", err)
				}
			},
		)
	}
	{
		g.Add(
			func() error {
				return worker.Start(workerCtx)
			},
			func(_ error) {
				workerCancel()
			},
		)
	}
	g.Add(run.SignalHandler(ctx, os.Interrupt))

	return g.Run()
}

func loadConfig() (*serviceConfig, error) {
	environment := os.Getenv("BEACON_ENV")
	if strings.ToUpper(environment) == "LOCAL" {
		if err := godotenv.Load("cmd/controller/.env"); err != nil {
			return nil, err
		}
	}

	var cfg serviceConfig
	if err := env.ParseWithOptions(&cfg, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
