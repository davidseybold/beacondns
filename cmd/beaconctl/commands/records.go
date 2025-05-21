package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/davidseybold/beacondns/client"
	"github.com/spf13/cobra"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Manage DNS records",
	Long:  `Commands for managing DNS resource record sets in Beacon.`,
}

var listRecordsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resource record sets in a zone",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		response, err := c.ListResourceRecordSets(context.Background(), zoneID)
		if err != nil {
			return err
		}

		if len(response.ResourceRecordSets) == 0 {
			fmt.Println("No records found")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tTTL\tVALUES")
		fmt.Fprintln(w, "────\t────\t───\t──────")
		for _, rrset := range response.ResourceRecordSets {
			values := ""
			for i, record := range rrset.ResourceRecords {
				if i > 0 {
					values += ", "
				}
				values += record.Value
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", rrset.Name, rrset.Type, rrset.TTL, values)
		}
		w.Flush()

		return nil
	},
}

var createRecordCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new resource record set",
	Long: `Create a new resource record set in a zone.
Example: beaconctl records create --zone-id 123 --name www.example.com --type A --ttl 300 --value 192.0.2.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		recordType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		ttl, err := cmd.Flags().GetUint32("ttl")
		if err != nil {
			return err
		}

		value, err := cmd.Flags().GetString("value")
		if err != nil {
			return err
		}

		changes := []client.Change{
			{
				Action: "CREATE",
				ResourceRecordSet: client.ResourceRecordSet{
					Name: name,
					Type: recordType,
					TTL:  ttl,
					ResourceRecords: []client.ResourceRecord{
						{Value: value},
					},
				},
			},
		}

		c := client.New(config.Host)
		changeInfo, err := c.ChangeResourceRecordSets(context.Background(), zoneID, changes)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintln(w, "─────\t─────")
		fmt.Fprintf(w, "Zone ID\t%s\n", zoneID)
		fmt.Fprintf(w, "Name\t%s\n", name)
		fmt.Fprintf(w, "Type\t%s\n", recordType)
		fmt.Fprintf(w, "TTL\t%d\n", ttl)
		fmt.Fprintf(w, "Value\t%s\n", value)
		fmt.Fprintf(w, "Change ID\t%s\n", changeInfo.ID)
		w.Flush()

		return nil
	},
}

var deleteRecordCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a resource record set",
	Long: `Delete a resource record set from a zone.
Example: beaconctl records delete --zone-id 123 --name www.example.com --type A`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		recordType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		changes := []client.Change{
			{
				Action: "DELETE",
				ResourceRecordSet: client.ResourceRecordSet{
					Name: name,
					Type: recordType,
				},
			},
		}

		c := client.New(config.Host)
		changeInfo, err := c.ChangeResourceRecordSets(context.Background(), zoneID, changes)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "FIELD\tVALUE")
		fmt.Fprintln(w, "─────\t─────")
		fmt.Fprintf(w, "Zone ID\t%s\n", zoneID)
		fmt.Fprintf(w, "Name\t%s\n", name)
		fmt.Fprintf(w, "Type\t%s\n", recordType)
		fmt.Fprintf(w, "Change ID\t%s\n", changeInfo.ID)
		w.Flush()

		return nil
	},
}

func init() {
	listRecordsCmd.Flags().String("zone-id", "", "ID of the zone")
	listRecordsCmd.MarkFlagRequired("zone-id")

	createRecordCmd.Flags().String("zone-id", "", "ID of the zone")
	createRecordCmd.Flags().String("name", "", "Name of the record (e.g., www.example.com)")
	createRecordCmd.Flags().String("type", "", "Type of the record (e.g., A, AAAA, CNAME, MX)")
	createRecordCmd.Flags().Uint32("ttl", 300, "Time to live in seconds")
	createRecordCmd.Flags().String("value", "", "Value of the record")
	createRecordCmd.MarkFlagRequired("zone-id")
	createRecordCmd.MarkFlagRequired("name")
	createRecordCmd.MarkFlagRequired("type")
	createRecordCmd.MarkFlagRequired("value")

	deleteRecordCmd.Flags().String("zone-id", "", "ID of the zone")
	deleteRecordCmd.Flags().String("name", "", "Name of the record")
	deleteRecordCmd.Flags().String("type", "", "Type of the record")
	deleteRecordCmd.MarkFlagRequired("zone-id")
	deleteRecordCmd.MarkFlagRequired("name")
	deleteRecordCmd.MarkFlagRequired("type")

	recordsCmd.AddCommand(listRecordsCmd, createRecordCmd, deleteRecordCmd)
	rootCmd.AddCommand(recordsCmd)
}
