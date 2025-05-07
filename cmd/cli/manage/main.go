package main

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type adminConfig struct {
	DBHost     string `env:"BEACON_DB_HOST"`
	DBName     string `env:"BEACON_DB_NAME" envDefault:"beacon_db"`
	DBUser     string `env:"BEACON_DB_USER" envDefault:"beacon_admin"`
	DBPass     string `env:"BEACON_DB_PASSWORD"`
	DBPort     int    `env:"BEACON_DB_PORT" envDefault:"5432"`
	Migrations string `env:"BEACON_MIGRATIONS_DIR" envDefault:"migrations"`
}

var cfg adminConfig

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load environment variables in local development
	if os.Getenv("BEACON_ENV") == "LOCAL" {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}

	if err := env.ParseWithOptions(&cfg, env.Options{
		RequiredIfNoDef: true,
	}); err != nil {
		return fmt.Errorf("error parsing config: %w", err)
	}

	rootCmd := &cobra.Command{
		Use:   "beacon-manage",
		Short: "BeaconDNS management CLI",
		Long:  `A command line tool for managing BeaconDNS system components such as database migrations and system configuration.`,
	}

	// Migration commands
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Long:  `Run all pending database migrations in the specified migrations directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleMigrate(cmd.Context(), cfg)
		},
	}

	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback the last migration",
		Long:  `Rollback the most recently applied database migration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleRollback(cmd.Context(), cfg)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check migration status",
		Long:  `Display the current status of all database migrations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleStatus(cmd.Context(), cfg)
		},
	}

	// Add commands to root
	rootCmd.AddCommand(migrateCmd, rollbackCmd, statusCmd)

	return rootCmd.Execute()
}

func handleMigrate(ctx context.Context, cfg adminConfig) error {
	// TODO: Implement database migration
	return fmt.Errorf("migrate command not implemented")
}

func handleRollback(ctx context.Context, cfg adminConfig) error {
	// TODO: Implement rollback
	return fmt.Errorf("rollback command not implemented")
}

func handleStatus(ctx context.Context, cfg adminConfig) error {
	// TODO: Implement status check
	return fmt.Errorf("status command not implemented")
}
