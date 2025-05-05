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
		Short: "Version of tmpltr you have installed",
		Long:  `Version of tmpltr you have installed`,

		RunE: func(c *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintf(c.OutOrStderr(), "%s\n", version)
			return nil
		},
	}

	return VersionCmd
}
