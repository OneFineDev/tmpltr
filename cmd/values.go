/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewValuesCommand() *cobra.Command {
	ValuesCmd := &cobra.Command{
		Use:   "values",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:
	
	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly Values a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceCmdCfg.SourceConfigFilePath = globalCfg.SourceConfigFile
			sourceCmdCfg.OutputPath = "temp"

			ss := services.NewSourceService(sourceCmdCfg, appLogger, cmd.Name())
			ss.Logger.Info("values called")

			err := ss.BuildProjectSourceConfigs()
			if err != nil {
				return err
			}
			ctx := context.Background()

			memFs := afero.NewMemMapFs()

			ss.CloneSources(ctx, memFs)

			ts := services.NewTemplateService(memFs)

			// Template handling
			err = ts.GetTemplateFiles(sourceCmdCfg.OutputPath)
			if err != nil {
				return err
			}
			err = ts.ParseTemplates()
			if err != nil {
				return err
			}

			// Values population
			ts.CreateTemplateValuesMap()

			out := cmd.OutOrStdout()

			p, err := yaml.Marshal(ts.TemplateValuesMap)
			if err != nil {
				ss.Logger.Error(err.Error())
				return err
			}

			out.Write(p)

			return nil
		},
	}

	ValuesCmd.Flags().StringVar(
		&sourceCmdCfg.SourceSet, "source-set", "", "the source set (defined in the sources config file) this execution will build",
	)
	ValuesCmd.Flags().StringSliceVar(
		&sourceCmdCfg.Sources, "sources", []string{}, "list of sources (defined in the sources config file) this execution will build",
	)

	ValuesCmd.MarkFlagsOneRequired("source-set", "sources")
	ValuesCmd.MarkFlagsMutuallyExclusive("source-set", "sources")

	return ValuesCmd
}
