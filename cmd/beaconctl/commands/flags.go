package commands

import "github.com/spf13/cobra"

const (
	defaultTTL = 300
)

type flagFunc func(cmd *cobra.Command)

func addFlags(flags []flagFunc, cmd *cobra.Command) {
	for _, flag := range flags {
		flag(cmd)
	}
}

func zoneIDFlag(required bool) flagFunc {
	return func(cmd *cobra.Command) {
		cmd.Flags().StringP("zone-id", "z", "", "ID of the zone")
		if required {
			_ = cmd.MarkFlagRequired("zone-id")
		}
	}
}

func typeFlag(required bool) flagFunc {
	return func(cmd *cobra.Command) {
		cmd.Flags().String("type", "", "Type of the record (e.g., A, AAAA, CNAME, MX)")
		if required {
			_ = cmd.MarkFlagRequired("type")
		}
	}
}

func ttlFlag(required bool) flagFunc {
	return func(cmd *cobra.Command) {
		cmd.Flags().Uint32("ttl", defaultTTL, "Time to live in seconds")
		if required {
			_ = cmd.MarkFlagRequired("ttl")
		}
	}
}

func valueFlag(required bool) flagFunc {
	return func(cmd *cobra.Command) {
		cmd.Flags().String("value", "", "Value of the record")
		if required {
			_ = cmd.MarkFlagRequired("value")
		}
	}
}
