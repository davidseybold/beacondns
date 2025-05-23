package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type Config struct {
	Host string `json:"host"`
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize beaconctl configuration",
	Long:  `Initialize beaconctl by setting the host URL for the Beacon DNS API.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		host, err := cmd.Flags().GetString("host")
		if err != nil {
			return err
		}

		if host == "" {
			return errors.New("host URL is required")
		}

		configDir, err := getConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config directory: %w", err)
		}

		config := Config{
			Host: host,
		}

		configFile := filepath.Join(configDir, "config")
		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err = os.WriteFile(configFile, data, 0600); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		cmd.Printf("Configuration saved to %s\n", configFile)
		return nil
	},
}

func init() {
	initCmd.Flags().String("host", "", "Host URL for the Beacon DNS API (e.g., http://localhost:8080)")
	_ = initCmd.MarkFlagRequired("host")
	rootCmd.AddCommand(initCmd)
}

func loadConfig() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	configFile := filepath.Join(configDir, "config")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
