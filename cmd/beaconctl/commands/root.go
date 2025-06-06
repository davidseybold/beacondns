package commands

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "beaconctl",
	Short: "Beaconctl is a CLI tool for managing Beacon DNS",
	Long: `Beaconctl is a command-line interface for managing Beacon DNS.
It allows you to create and manage DNS zones and resource record sets.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configDir := filepath.Join(home, ".beacon")
	if err = os.MkdirAll(configDir, 0700); err != nil {
		return "", err
	}
	return configDir, nil
}
