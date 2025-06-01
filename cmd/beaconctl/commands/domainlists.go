package commands

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/davidseybold/beacondns/client"
	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var domainListsCmd = &cobra.Command{
	Use:   "domain-lists",
	Short: "Manage domain lists",
	Long:  `Commands for managing domain lists in Beacon.`,
}

var createDomainListCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new domain list",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		domains, err := cmd.Flags().GetStringSlice("domains")
		if err != nil {
			return err
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		if file != "" {
			fileDomains, err := readDomainsFromFile(file)
			if err != nil {
				return err
			}
			domains = append(domains, fileDomains...)
		}

		c := client.New(config.Host)
		domainList, err := c.CreateDomainList(context.Background(), client.CreateDomainListRequest{
			Name:    name,
			Domains: domains,
		})
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "DOMAIN COUNT"})
		_ = table.Append([]string{domainList.ID.String(), domainList.Name, strconv.Itoa(domainList.DomainCount)})
		return table.Render()
	},
}

func readDomainsFromFile(file string) ([]string, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n"), nil
}

var deleteDomainListCmd = &cobra.Command{
	Use:   "delete [domain-list-id]",
	Short: "Delete a domain list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		domainListID, err := uuid.Parse(args[0])
		if err != nil {
			cmd.PrintErrln("invalid domain list ID")
			return err
		}

		c := client.New(config.Host)
		err = c.DeleteDomainList(context.Background(), domainListID)
		if err != nil {
			cmd.PrintErrf("failed to delete domain list: %v", err)
			return err
		}

		cmd.Println("Domain list deleted")
		return nil
	},
}

var listDomainListsCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domain lists",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		domainLists, err := c.GetDomainLists(context.Background())
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "DOMAIN COUNT"})
		for _, domainList := range domainLists {
			table.Append([]string{domainList.ID.String(), domainList.Name, strconv.Itoa(domainList.DomainCount)})
		}
		return table.Render()
	},
}

var getDomainListCmd = &cobra.Command{
	Use:   "get [domain-list-id]",
	Short: "Get a domain list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		domainListID, err := uuid.Parse(args[0])
		if err != nil {
			cmd.PrintErrln("invalid domain list ID")
			return err
		}

		c := client.New(config.Host)
		domainList, err := c.GetDomainList(context.Background(), domainListID)
		if err != nil {
			cmd.PrintErrf("failed to get domain list: %v", err)
			return err
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"ID", "NAME", "DOMAIN COUNT"})
		_ = table.Append([]string{domainList.ID.String(), domainList.Name, strconv.Itoa(domainList.DomainCount)})
		return table.Render()
	},
}

func init() {
	createDomainListCmd.Flags().StringP("name", "n", "", "Name of the domain list")
	createDomainListCmd.Flags().StringSliceP("domains", "d", []string{}, "Domains to add to the domain list")
	createDomainListCmd.Flags().StringP("file", "f", "", "File to read domains from")
	_ = createDomainListCmd.MarkFlagRequired("name")

	domainListsCmd.AddCommand(createDomainListCmd, deleteDomainListCmd, getDomainListCmd, listDomainListsCmd)
	rootCmd.AddCommand(domainListsCmd)
}
