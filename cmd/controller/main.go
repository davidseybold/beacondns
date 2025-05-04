package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/davidseybold/beacondns/internal/controller"
)

type serviceConfig struct {
	Port   int    `env:"BEACON_PORT" envDefault:"8080"`
	DBHost string `env:"BEACON_DB_HOST"`
	DBName string `env:"BEACON_DB_NAME" envDefault:"beacon_db"`
	DBUser string `env:"BEACON_DB_USER" envDefault:"beacon_controller"`
	DBPass string `env:"BEACON_DB_PASSWORD"`
	DBPort int    `env:"BEACON_DB_PORT" envDefault:"5432"`
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

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

	controller, err := controller.NewServer(ctx, controller.ControllerSettings{
		Port:       cfg.Port,
		DBHost:     cfg.DBHost,
		DBName:     cfg.DBName,
		DBUser:     cfg.DBUser,
		DBPassword: cfg.DBPass,
		DBPort:     cfg.DBPort,
	})

	if err != nil {
		return err
	}

	if err := controller.Start(ctx); err != nil {
		return err
	}

	return nil
}
