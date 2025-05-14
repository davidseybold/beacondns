package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/oklog/run"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/davidseybold/beacondns/internal/logger"
	"github.com/davidseybold/beacondns/internal/messaging"
	"github.com/davidseybold/beacondns/internal/resolver"
)

const (
	exchangeName = "beacon"
)

type config struct {
	RabbitHost string `env:"BEACON_RABBITMQ_HOST"`
	Host       string `env:"BEACON_HOST"`
}

func (c *config) Validate() error {
	return nil
}

func main() {
	ctx := context.Background()
	if err := start(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func start(ctx context.Context, _ io.Writer) error {
	logger := logger.NewJSONLogger(slog.LevelInfo, os.Stdout)

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	publishConn, err := amqp.Dial(cfg.RabbitHost)
	if err != nil {
		return fmt.Errorf("error creating RabbitMQ connection: %w", err)
	}
	defer publishConn.Close()

	publisher, err := messaging.NewRabbitMQPublisher(publishConn, "beacon")
	if err != nil {
		return fmt.Errorf("error creating RabbitMQ publisher: %w", err)
	}
	defer publisher.Close()

	consumeConn, err := amqp.Dial(cfg.RabbitHost)
	if err != nil {
		return fmt.Errorf("error creating RabbitMQ connection: %w", err)
	}
	defer consumeConn.Close()

	queueName := fmt.Sprintf("server.resolver.%s", cfg.Host)
	consumerName := fmt.Sprintf("resolver.%s", cfg.Host)

	consumer := messaging.NewRabbitMQConsumer(consumerName, consumeConn)

	err = messaging.SetupRabbitMQTopology(consumeConn, messaging.RabbitMQTopology{
		Exchange: messaging.RabbitMQExchange{
			Name: exchangeName,
			Kind: "topic",
		},
		Queues: []string{
			queueName,
		},
	})
	if err != nil {
		return fmt.Errorf("error setting up RabbitMQ topology: %w", err)
	}

	clCtx, clCancel := context.WithCancel(ctx)
	changeListener := resolver.NewChangeListener(clCtx, resolver.ChangeListenerConfig{
		Consumer:    consumer,
		Publisher:   publisher,
		ChangeQueue: queueName,
		Logger:      logger,
	})

	var g run.Group
	{
		g.Add(func() error {
			return changeListener.Run()
		}, func(_ error) {
			clCancel()
		})
	}
	g.Add(run.SignalHandler(ctx, os.Interrupt))

	return g.Run()
}

func loadConfig() (*config, error) {
	environment := os.Getenv("BEACON_ENV")
	if strings.ToUpper(environment) == "LOCAL" {
		if err := godotenv.Load("cmd/resolver/.env"); err != nil {
			return nil, err
		}
	}

	var cfg config
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
