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
