package commands

import (
	"context"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/davidseybold/beacondns/client"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage DNS zones",
	Long:  `Commands for managing DNS zones in Beacon.`,
}

var createZoneCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new DNS zone",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		name := args[0]

		c := client.New(config.Host)
		response, err := c.CreateZone(context.Background(), name)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "RECORD COUNT"})
		_ = table.Append(
			[]string{response.Zone.ID, response.Zone.Name, strconv.Itoa(response.Zone.ResourceRecordSetCount)},
		)
		return table.Render()
	},
}

var listZonesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DNS zones",
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		response, err := c.ListZones(context.Background())
		if err != nil {
			return err
		}

		if len(response.Zones) == 0 {
			cmd.Println("No zones found")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "RECORD COUNT"})
		for _, zone := range response.Zones {
			_ = table.Append([]string{zone.ID, zone.Name, strconv.Itoa(zone.ResourceRecordSetCount)})
		}
		return table.Render()
	},
}

var describeZoneCmd = &cobra.Command{
	Use:   "describe [zone-id]",
	Short: "Get information about a specific zone",
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		zone, err := c.GetZone(context.Background(), zoneID)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "RECORD COUNT"})
		_ = table.Append([]string{zone.ID, zone.Name, strconv.Itoa(zone.ResourceRecordSetCount)})
		return table.Render()
	},
}

var deleteZoneCmd = &cobra.Command{
	Use:   "delete [zone-id]",
	Short: "Delete a DNS zone",
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("id")
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		response, err := c.DeleteZone(context.Background(), zoneID)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ZONE ID", "CHANGE ID"})
		_ = table.Append([]string{zoneID, response.ID})
		return table.Render()
	},
}

func init() {
	zonesCmd.AddCommand(createZoneCmd, listZonesCmd, describeZoneCmd, deleteZoneCmd)
	rootCmd.AddCommand(zonesCmd)
}
