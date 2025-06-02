package commands

import (
	"context"
	"strconv"

	"github.com/google/uuid"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/davidseybold/beacondns/client"
)

var firewallRulesCmd = &cobra.Command{
	Use:   "firewall-rules",
	Short: "Manage firewall rules",
	Long:  `Commands for managing firewall rules in Beacon.`,
}

var createFirewallRuleCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new firewall rule",
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		domainListIDFlagVal, err := cmd.Flags().GetString("domain-list-id")
		if err != nil {
			return err
		}

		domainListID, err := uuid.Parse(domainListIDFlagVal)
		if err != nil {
			cmd.PrintErrln("invalid domain list ID")
			return err
		}

		action, err := cmd.Flags().GetString("action")
		if err != nil {
			return err
		}

		priority, err := cmd.Flags().GetUint("priority")
		if err != nil {
			return err
		}

		blockResponseType, err := cmd.Flags().GetString("block-response-type")
		if err != nil {
			return err
		}

		blockResponseRecordTTL, err := cmd.Flags().GetUint32("block-response-record-ttl")
		if err != nil {
			return err
		}

		blockResponseRecordTypeFlagVal, err := cmd.Flags().GetString("block-response-record-type")
		if err != nil {
			return err
		}

		var blockResponseRecordType *string
		if blockResponseRecordTypeFlagVal != "" {
			blockResponseRecordType = &blockResponseRecordTypeFlagVal
		}

		blockResponseRecordValuesFlagVal, err := cmd.Flags().GetStringSlice("block-response-record-values")
		if err != nil {
			return err
		}

		blockResponseRecordValues := make([]client.ResourceRecord, len(blockResponseRecordValuesFlagVal))
		for i, value := range blockResponseRecordValuesFlagVal {
			blockResponseRecordValues[i] = client.ResourceRecord{Value: value}
		}

		var blockResponse *client.ResourceRecordSet
		if blockResponseType == "override" {
			blockResponse = &client.ResourceRecordSet{
				ResourceRecords: blockResponseRecordValues,
				TTL:             blockResponseRecordTTL,
				Type:            *blockResponseRecordType,
			}
		}

		c := client.New(config.Host)
		rule, err := c.CreateFirewallRule(context.Background(), client.CreateFirewallRuleRequest{
			Name:              name,
			DomainListID:      domainListID,
			Action:            action,
			BlockResponseType: &blockResponseType,
			BlockResponse:     blockResponse,
			Priority:          priority,
		})
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.Header([]string{"ID", "NAME", "DOMAIN LIST ID", "ACTION", "PRIORITY"})
		_ = table.Append(
			[]string{
				rule.ID.String(),
				rule.Name,
				rule.DomainListID.String(),
				rule.Action,
				strconv.Itoa(int(rule.Priority)),
			},
		)
		return table.Render()
	},
}

var deleteFirewallRuleCmd = &cobra.Command{
	Use:   "delete [firewall-rule-id]",
	Short: "Delete a firewall rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		ruleID, err := uuid.Parse(args[0])
		if err != nil {
			cmd.PrintErrln("invalid firewall rule ID")
			return err
		}

		c := client.New(config.Host)
		err = c.DeleteFirewallRule(context.Background(), ruleID)
		if err != nil {
			return err
		}

		cmd.Println("Firewall rule deleted successfully")
		return nil
	},
}

var getFirewallRuleCmd = &cobra.Command{
	Use:   "get [firewall-rule-id]",
	Short: "Get a firewall rule",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		ruleID, err := uuid.Parse(args[0])
		if err != nil {
			cmd.PrintErrln("invalid firewall rule ID")
			return err
		}

		c := client.New(config.Host)
		rule, err := c.GetFirewallRule(context.Background(), ruleID)
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.Header([]string{"ID", "NAME", "DOMAIN LIST ID", "ACTION", "PRIORITY"})
		_ = table.Append(
			[]string{
				rule.ID.String(),
				rule.Name,
				rule.DomainListID.String(),
				rule.Action,
				strconv.Itoa(int(rule.Priority)),
			},
		)
		return table.Render()
	},
}

var listFirewallRulesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all firewall rules",
	RunE: func(cmd *cobra.Command, _ []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		c := client.New(config.Host)
		rules, err := c.GetFirewallRules(context.Background())
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.Header([]string{"ID", "NAME", "DOMAIN LIST ID", "ACTION", "PRIORITY"})
		for _, rule := range rules {
			_ = table.Append(
				[]string{
					rule.ID.String(),
					rule.Name,
					rule.DomainListID.String(),
					rule.Action,
					strconv.Itoa(int(rule.Priority)),
				},
			)
		}
		return table.Render()
	},
}

var updateFirewallRuleCmd = &cobra.Command{
	Use:   "update [firewall-rule-id]",
	Short: "Update a firewall rule",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := loadConfig()
		if err != nil {
			return err
		}

		ruleID, err := uuid.Parse(args[0])
		if err != nil {
			cmd.PrintErrln("invalid firewall rule ID")
			return err
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}

		action, err := cmd.Flags().GetString("action")
		if err != nil {
			return err
		}

		priority, err := cmd.Flags().GetUint("priority")
		if err != nil {
			return err
		}

		blockResponseType, err := cmd.Flags().GetString("block-response-type")
		if err != nil {
			return err
		}

		blockResponseRecordTTL, err := cmd.Flags().GetUint32("block-response-record-ttl")
		if err != nil {
			return err
		}

		blockResponseRecordTypeFlagVal, err := cmd.Flags().GetString("block-response-record-type")
		if err != nil {
			return err
		}

		var blockResponseRecordType *string
		if blockResponseRecordTypeFlagVal != "" {
			blockResponseRecordType = &blockResponseRecordTypeFlagVal
		}

		blockResponseRecordValuesFlagVal, err := cmd.Flags().GetStringSlice("block-response-record-values")
		if err != nil {
			return err
		}

		blockResponseRecordValues := make([]client.ResourceRecord, len(blockResponseRecordValuesFlagVal))
		for i, value := range blockResponseRecordValuesFlagVal {
			blockResponseRecordValues[i] = client.ResourceRecord{Value: value}
		}

		var blockResponse *client.ResourceRecordSet
		if blockResponseType == "override" {
			blockResponse = &client.ResourceRecordSet{
				ResourceRecords: blockResponseRecordValues,
				TTL:             blockResponseRecordTTL,
				Type:            *blockResponseRecordType,
			}
		}

		c := client.New(config.Host)
		rule, err := c.UpdateFirewallRule(context.Background(), ruleID, client.UpdateFirewallRuleRequest{
			Name:              name,
			Action:            action,
			Priority:          priority,
			BlockResponseType: &blockResponseType,
			BlockResponse:     blockResponse,
		})
		if err != nil {
			return err
		}

		table := tablewriter.NewWriter(cmd.OutOrStdout())
		table.Header([]string{"ID", "NAME", "DOMAIN LIST ID", "ACTION", "PRIORITY"})
		_ = table.Append(
			[]string{
				rule.ID.String(),
				rule.Name,
				rule.DomainListID.String(),
				rule.Action,
				strconv.Itoa(int(rule.Priority)),
			},
		)
		return table.Render()
	},
}

func init() {
	createFirewallRuleCmd.Flags().String("name", "", "Name of the firewall rule")
	createFirewallRuleCmd.Flags().String("domain-list-id", "", "ID of the domain list to apply the firewall rule to")
	createFirewallRuleCmd.Flags().
		String("action", "", "Action to take when the firewall rule is triggered (block, allow, alert)")
	createFirewallRuleCmd.Flags().Uint("priority", 0, "Priority of the firewall rule")
	createFirewallRuleCmd.Flags().
		String("block-response-type", "", "Type of response to return when the firewall rule is triggered (nxdomain, nodata, override)")
	createFirewallRuleCmd.Flags().
		StringSlice("block-response-record-values", []string{}, "Values to return when the firewall rule is triggered (nxdomain, nodata, override)")
	createFirewallRuleCmd.Flags().
		Uint32("block-response-record-ttl", 0, "TTL to return when the firewall rule is triggered (nxdomain, nodata, override)")
	createFirewallRuleCmd.Flags().
		String("block-response-record-type", "", "Type of response to return when the firewall rule is triggered (A, AAAA, CNAME)")
	_ = createFirewallRuleCmd.MarkFlagRequired("name")
	_ = createFirewallRuleCmd.MarkFlagRequired("domain-list-id")
	_ = createFirewallRuleCmd.MarkFlagRequired("action")
	_ = createFirewallRuleCmd.MarkFlagRequired("priority")
	_ = createFirewallRuleCmd.MarkFlagRequired("block-response-type")

	updateFirewallRuleCmd.Flags().String("name", "", "Name of the firewall rule")
	updateFirewallRuleCmd.Flags().
		String("action", "", "Action to take when the firewall rule is triggered (block, allow, alert)")
	updateFirewallRuleCmd.Flags().Uint("priority", 0, "Priority of the firewall rule")
	updateFirewallRuleCmd.Flags().
		String("block-response-type", "", "Type of response to return when the firewall rule is triggered (nxdomain, nodata, override)")
	updateFirewallRuleCmd.Flags().
		StringSlice("block-response-record-values", []string{}, "Values to return when the firewall rule is triggered (nxdomain, nodata, override)")
	updateFirewallRuleCmd.Flags().
		Uint32("block-response-record-ttl", 0, "TTL to return when the firewall rule is triggered (nxdomain, nodata, override)")
	updateFirewallRuleCmd.Flags().
		String("block-response-record-type", "", "Type of response to return when the firewall rule is triggered (A, AAAA, CNAME)")

	_ = updateFirewallRuleCmd.MarkFlagRequired("name")
	_ = updateFirewallRuleCmd.MarkFlagRequired("action")
	_ = updateFirewallRuleCmd.MarkFlagRequired("priority")
	_ = updateFirewallRuleCmd.MarkFlagRequired("block-response-type")

	firewallRulesCmd.AddCommand(
		createFirewallRuleCmd,
		deleteFirewallRuleCmd,
		getFirewallRuleCmd,
		listFirewallRulesCmd,
		updateFirewallRuleCmd,
	)
	rootCmd.AddCommand(firewallRulesCmd)
}
