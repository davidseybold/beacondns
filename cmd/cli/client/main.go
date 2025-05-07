package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type configFile struct {
	ControllerURL string `yaml:"controller_url"`
	APIKey        string `yaml:"api_key"`
}

var (
	cfg        configFile
	configPath string
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Set default config path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	configPath = filepath.Join(homeDir, ".beacon", "config.yaml")

	// Load environment variables in local development
	if os.Getenv("BEACON_ENV") == "LOCAL" {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}

	// Load config from file
	if err := loadConfigFromFile(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error loading config file: %w", err)
		}
		// If config file doesn't exist, initialize with empty values
		cfg = configFile{}
	}

	rootCmd := &cobra.Command{
		Use:   "beacon",
		Short: "BeaconDNS client CLI",
		Long:  `A command line tool for interacting with the BeaconDNS controller API.`,
	}

	// Config command
	configCmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure the CLI",
		Long:  `Configure the CLI by setting the controller URL and API key.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			controllerURL, _ := cmd.Flags().GetString("controller-url")
			apiKey, _ := cmd.Flags().GetString("api-key")

			if controllerURL == "" && apiKey == "" {
				return fmt.Errorf("at least one of --controller-url or --api-key must be provided")
			}

			// Update config with new values
			if controllerURL != "" {
				cfg.ControllerURL = controllerURL
			}
			if apiKey != "" {
				cfg.APIKey = apiKey
			}

			// Create config directory if it doesn't exist
			configDir := filepath.Dir(configPath)
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			// Write config to file
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(configPath, data, 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Println("Configuration saved successfully")
			return nil
		},
	}

	// Add flags to config command
	configCmd.Flags().String("controller-url", "", "Controller API URL")
	configCmd.Flags().String("api-key", "", "API key for authentication")

	// Resource commands
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List resources",
		Long:  `List all resources of the specified type.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleList(cmd.Context())
		},
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new resource",
		Long:  `Create a new resource with the specified parameters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleCreate(cmd.Context())
		},
	}

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a resource",
		Long:  `Delete an existing resource by its ID.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleDelete(cmd.Context())
		},
	}

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a resource",
		Long:  `Update an existing resource with new parameters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handleUpdate(cmd.Context())
		},
	}

	// Add commands to root
	rootCmd.AddCommand(configCmd, listCmd, createCmd, deleteCmd, updateCmd)

	return rootCmd.Execute()
}

func loadConfigFromFile() error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func handleList(ctx context.Context) error {
	if cfg.ControllerURL == "" || cfg.APIKey == "" {
		return fmt.Errorf("configuration is incomplete. Please run 'beacon-client config' to set up the CLI")
	}
	// TODO: Implement list command
	return fmt.Errorf("list command not implemented")
}

func handleCreate(ctx context.Context) error {
	if cfg.ControllerURL == "" || cfg.APIKey == "" {
		return fmt.Errorf("configuration is incomplete. Please run 'beacon-client config' to set up the CLI")
	}
	// TODO: Implement create command
	return fmt.Errorf("create command not implemented")
}

func handleDelete(ctx context.Context) error {
	if cfg.ControllerURL == "" || cfg.APIKey == "" {
		return fmt.Errorf("configuration is incomplete. Please run 'beacon-client config' to set up the CLI")
	}
	// TODO: Implement delete command
	return fmt.Errorf("delete command not implemented")
}

func handleUpdate(ctx context.Context) error {
	if cfg.ControllerURL == "" || cfg.APIKey == "" {
		return fmt.Errorf("configuration is incomplete. Please run 'beacon-client config' to set up the CLI")
	}
	// TODO: Implement update command
	return fmt.Errorf("update command not implemented")
}
