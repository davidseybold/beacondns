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

type config struct {
	EtcdEndpoints []string      `env:"BEACON_ETCD_ENDPOINTS" envSeparator:","`
	ResolverType  resolver.Type `env:"BEACON_RESOLVER_TYPE"`
	Forwarder     string        `env:"BEACON_FORWARDER"`
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

	resolverCtx, cancelResolver := context.WithCancel(ctx)
	dnsresolver, err := resolver.New(&resolver.Config{
		Type:          cfg.ResolverType,
		Forwarder:     &cfg.Forwarder,
		EtcdEndpoints: cfg.EtcdEndpoints,
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
