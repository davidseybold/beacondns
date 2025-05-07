package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/davidseybold/beacondns/internal/controller/api"
	"github.com/davidseybold/beacondns/internal/controller/outbox"
	"github.com/davidseybold/beacondns/internal/controller/repository"
	"github.com/davidseybold/beacondns/internal/controller/usecase"
	"github.com/davidseybold/beacondns/internal/libs/db/postgres"
	"github.com/davidseybold/beacondns/internal/libs/messaging"
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
	RabbitHost             string `env:"BEACON_RABBITMQ_HOST"`
	ShutdownTimeout        int    `env:"BEACON_SHUTDOWN_TIMEOUT" envDefault:"30"`
}

func (c *serviceConfig) Validate() error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}
	if c.DBPort <= 0 || c.DBPort > 65535 {
		return fmt.Errorf("invalid database port number: %d", c.DBPort)
	}
	if c.OutboxBatchSize <= 0 {
		return fmt.Errorf("invalid outbox batch size: %d", c.OutboxBatchSize)
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

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
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

	publisherConn, err := amqp.Dial(cfg.RabbitHost)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer publisherConn.Close()

	repoRegistry := repository.NewPostgresRepositoryRegistry(db)
	publisher, err := messaging.NewRabbitMQPublisher(publisherConn)
	if err != nil {
		return fmt.Errorf("failed to create publisher: %w", err)
	}

	controllerService := usecase.NewControllerService(repoRegistry)
	outboxService := usecase.NewOutboxService(repoRegistry, publisher)

	outboxCtx, cancelOutbox := context.WithCancel(ctx)
	outboxProcessor := outbox.NewProcessor(outboxCtx, outboxService, cfg.OutboxBatchSize)

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
				shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ShutdownTimeout)*time.Second)
				defer cancel()
				if err := httpServer.Shutdown(shutdownCtx); err != nil {
					fmt.Fprintf(w, "error shutting down HTTP server: %s\n", err)
				}
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
