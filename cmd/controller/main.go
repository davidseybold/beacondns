package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/oklog/run"

	"github.com/davidseybold/beacondns/internal/controller/api"
	"github.com/davidseybold/beacondns/internal/controller/outbox"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/controller/usecase"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
)

type serviceConfig struct {
	Port                   int    `env:"BEACON_CONTROLLER_PORT" envDefault:"8080"`
	DBHost                 string `env:"BEACON_DB_HOST"`
	DBName                 string `env:"BEACON_DB_NAME" envDefault:"beacon_db"`
	DBUser                 string `env:"BEACON_DB_USER" envDefault:"beacon_controller"`
	DBPass                 string `env:"BEACON_DB_PASSWORD"`
	DBPort                 int    `env:"BEACON_DB_PORT" envDefault:"5432"`
	OutboxProcessorEnabled bool   `env:"BEACON_OUTBOX_PROCESSOR_ENABLED" envDefault:"true"`
	OutboxBatchSize        int    `env:"BEACON_OUTBOX_BATCH_SIZE" envDefault:"10"`
}

func main() {
	ctx := context.Background()
	if err := start(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func start(ctx context.Context, w io.Writer) error {

	environment := os.Getenv("BEACON_ENV")
	if strings.ToUpper(environment) == "LOCAL" {
		if err := godotenv.Load(); err != nil {
			return err
		}
	}

	var cfg serviceConfig
	if err := env.ParseWithOptions(&cfg, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		return err
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

	controllerService := usecase.NewControllerService(repoRegistry)
	outboxService := usecase.NewOutboxService()

	outboxCtx, cancelOutbox := context.WithCancel(ctx)

	outboxProcessor := outbox.NewProcessor(outboxCtx, repoRegistry, outboxService, cfg.OutboxBatchSize)

	var g run.Group
	{
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: api.NewHTTPHandler(controllerService),
		}
		g.Add(
			func() error {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					return err
				}
				return nil
			},
			func(_ error) {
				httpServer.Shutdown(context.Background())
			},
		)
	}
	{
		g.Add(
			func() error {
				return outboxProcessor.Run()
			},
			func(_ error) {
				cancelOutbox()
			},
		)
	}

	g.Add(run.SignalHandler(ctx, os.Interrupt))

	return g.Run()
}
