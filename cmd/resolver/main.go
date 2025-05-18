package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/oklog/run"

	"github.com/davidseybold/beacondns/internal/resolver"
)

const (
	exchangeName = "beacon"
)

type config struct {
	DBPath       string        `env:"BEACON_DB_PATH"       envDefault:"/var/lib/beacon"`
	RabbitHost   string        `env:"BEACON_RABBITMQ_HOST"`
	Host         string        `env:"BEACON_HOSTNAME"`
	ResolverType resolver.Type `env:"BEACON_RESOLVER_TYPE"`
	Forwarder    string        `env:"BEACON_FORWARDER"`
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
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	queueName := fmt.Sprintf("server.resolver.%s", cfg.Host)

	resolverCtx, cancelResolver := context.WithCancel(ctx)
	dnsresolver, err := resolver.New(&resolver.Config{
		Type:               cfg.ResolverType,
		Forwarder:          &cfg.Forwarder,
		HostName:           cfg.Host,
		DBPath:             cfg.DBPath,
		RabbitMQConnString: cfg.RabbitHost,
		RabbitExchange:     exchangeName,
		ChangeQueue:        queueName,
	})
	if err != nil {
		cancelResolver()
		return err
	}

	var g run.Group
	{
		g.Add(func() error {
			return dnsresolver.Run(resolverCtx)
		}, func(_ error) {
			cancelResolver()
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
