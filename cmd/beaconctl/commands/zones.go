package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/davidseybold/beacondns/client"
	"github.com/spf13/cobra"
)

var zonesCmd = &cobra.Command{
	Use:   "zones",
	Short: "Manage DNS zones",
	Long:  `Commands for managing DNS zones in Beacon.`,
}

var createZoneCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new DNS zone",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		response, err := c.CreateZone(context.Background(), name)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tRECORD COUNT")
		fmt.Fprintf(w, "%s\t%s\t%d\n", response.Zone.ID, response.Zone.Name, response.Zone.ResourceRecordSetCount)
		w.Flush()
		return nil
	},
}

var listZonesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DNS zones",
	RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Println("No zones found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tRECORD COUNT")
		fmt.Fprintln(w, "──\t────\t───────────")
		for _, zone := range response.Zones {
			fmt.Fprintf(w, "%s\t%s\t%d\n", zone.ID, zone.Name, zone.ResourceRecordSetCount)
		}
		w.Flush()

		return nil
	},
}

var getZoneCmd = &cobra.Command{
	Use:   "get",
	Short: "Get information about a specific zone",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintln(w, "─────\t─────")
		fmt.Fprintf(w, "ID\t%s\n", zone.ID)
		fmt.Fprintf(w, "Name\t%s\n", zone.Name)
		fmt.Fprintf(w, "Record Count\t%d\n", zone.ResourceRecordSetCount)
		w.Flush()

		return nil
	},
}

var deleteZoneCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a DNS zone",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintln(w, "─────\t─────")
		fmt.Fprintf(w, "Zone ID\t%s\n", zoneID)
		fmt.Fprintf(w, "Change ID\t%s\n", response.ID)
		w.Flush()

		return nil
	},
}

func init() {
	createZoneCmd.Flags().String("name", "", "Name of the zone (e.g., example.com)")
	createZoneCmd.MarkFlagRequired("name")

	getZoneCmd.Flags().String("id", "", "ID of the zone")
	getZoneCmd.MarkFlagRequired("id")

	deleteZoneCmd.Flags().String("id", "", "ID of the zone")
	deleteZoneCmd.MarkFlagRequired("id")

	zonesCmd.AddCommand(createZoneCmd, listZonesCmd, getZoneCmd, deleteZoneCmd)
	rootCmd.AddCommand(zonesCmd)
}
