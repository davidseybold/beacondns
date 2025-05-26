package commands

import (
	"context"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/davidseybold/beacondns/client"
)

var recordsCmd = &cobra.Command{
	Use:   "record-sets",
	Short: "Manage DNS records",
	Long:  `Commands for managing DNS resource record sets in Beacon.`,
}

var listRecordsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all resource record sets in a zone",
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		rrSets, err := c.ListResourceRecordSets(context.Background(), zoneID)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"NAME", "TYPE", "TTL", "VALUES"})
		for _, rrset := range rrSets {
			values := ""
			for i, record := range rrset.ResourceRecords {
				if i > 0 {
					values += ", "
				}
				values += record.Value
			}
			_ = table.Append([]string{rrset.Name, rrset.Type, strconv.Itoa(int(rrset.TTL)), values})
		}
		return table.Render()
	},
}

var createRecordCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new resource record set",
	Long: `Create a new resource record set in a zone.
Example: beaconctl records create www.example.com --zone-id 123 --type A --ttl 300 --value 192.0.2.1`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
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

		name := args[0]

		c := client.New(config.Host)
		rrSet, err := c.UpsertResourceRecordSet(context.Background(), zoneID, client.ResourceRecordSet{
			Name: name,
			Type: recordType,
			TTL:  ttl,
			ResourceRecords: []client.ResourceRecord{
				{Value: value},
			},
		})
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ZONE ID", "NAME", "TYPE", "TTL", "VALUE"})
		_ = table.Append([]string{zoneID, rrSet.Name, rrSet.Type, strconv.Itoa(int(rrSet.TTL)), value})
		return table.Render()
	},
}

var deleteRecordCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a resource record set",
	Long: `Delete a resource record set from a zone.
Example: beaconctl records delete www.example.com --zone-id 123 --type A`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
		if err != nil {
			return err
		}

		recordType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		name := args[0]

		c := client.New(config.Host)
		err = c.DeleteResourceRecordSet(context.Background(), zoneID, name, recordType)
		if err != nil {
			return err
		}

		cmd.Println("Record set deleted")
		return nil
	},
}

var getRecordCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a resource record set",
	Long: `Get a resource record set from a zone.
Example: beaconctl records get www.example.com --zone-id 123 --type A`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		zoneID, err := cmd.Flags().GetString("zone-id")
		if err != nil {
			return err
		}

		recordType, err := cmd.Flags().GetString("type")
		if err != nil {
			return err
		}

		name := args[0]

		c := client.New(config.Host)
		rrSet, err := c.GetResourceRecordSet(context.Background(), zoneID, name, recordType)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"NAME", "TYPE", "TTL", "VALUES"})
		_ = table.Append([]string{rrSet.Name, rrSet.Type, strconv.Itoa(int(rrSet.TTL)), rrSet.ResourceRecords[0].Value})
		return table.Render()
	},
}

func init() {
	listRecordsFlags := []flagFunc{
		zoneIDFlag(),
	}

	createRecordFlags := []flagFunc{
		zoneIDFlag(),
		typeFlag(true),
		ttlFlag(false),
		valueFlag(true),
	}

	deleteRecordFlags := []flagFunc{
		zoneIDFlag(),
		typeFlag(true),
	}

	getRecordFlags := []flagFunc{
		zoneIDFlag(),
		typeFlag(true),
	}

	addFlags(listRecordsFlags, listRecordsCmd)
	addFlags(createRecordFlags, createRecordCmd)
	addFlags(deleteRecordFlags, deleteRecordCmd)
	addFlags(getRecordFlags, getRecordCmd)

	recordsCmd.AddCommand(listRecordsCmd, createRecordCmd, deleteRecordCmd, getRecordCmd)
	rootCmd.AddCommand(recordsCmd)
}
