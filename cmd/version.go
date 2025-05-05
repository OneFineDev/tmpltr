package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
)

func NewVersionCommand() *cobra.Command {
	VersionCmd := &cobra.Command{
		Use:   "version",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly Version a Cobra application.`,

		RunE: func(c *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(c.OutOrStderr(), "%s\n", version)
			return nil
		},
	}

	return VersionCmd
}
