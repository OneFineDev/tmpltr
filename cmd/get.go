package cmd

import (
	"github.com/spf13/cobra"
)

func NewGetCommand() *cobra.Command {
	GetCmd := &cobra.Command{
		Use:   "get",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly Get a Cobra application.`,
		Run: func(_ *cobra.Command, _ []string) {},
	}

	GetCmd.AddCommand(
		NewValuesCommand(),
	)
	return GetCmd
}
