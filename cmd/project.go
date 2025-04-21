/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var sourceCmdCfg *services.SourcesCommandConfig = &services.SourcesCommandConfig{}

func NewProjectCommand() *cobra.Command {
	ProjectCmd := &cobra.Command{

		Use:   "project",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement logging
			sourceCmdCfg.SourceConfigFilePath = globalCfg.SourceConfigFile

			ss := services.NewSourceService(sourceCmdCfg, appLogger, cmd.Name())
			err := ss.BuildProjectSourceConfigs()
			if err != nil {
				return err
			}

			osFs := afero.NewOsFs()

			ctx := context.Background()
			ss.CloneSources(ctx, osFs)

			ts := services.NewTemplateService(osFs)

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

			ts.InteractiveInput()

			return nil
		},
	}

	ProjectCmd.Flags().StringVarP(
		&sourceCmdCfg.OutputPath, "output-path", "o", "", "the path in which the sourceset/sources will be rendered",
	)
	ProjectCmd.Flags().StringVarP(
		&sourceCmdCfg.ProjectName, "project-name", "p", "", "name of the project/new repository",
	)
	ProjectCmd.Flags().BoolVarP(
		&sourceCmdCfg.AppendProjectName, "append-name", "a", true, "whether to append this project name to the output path so the project gets built in output-path/project-name",
	)
	ProjectCmd.Flags().StringVar(
		&sourceCmdCfg.SourceSet, "source-set", "", "the source set (defined in the sources config file) this execution will build",
	)
	ProjectCmd.Flags().StringSliceVar(
		&sourceCmdCfg.Sources, "sources", []string{}, "list of sources (defined in the sources config file) this execution will build",
	)

	ProjectCmd.MarkFlagRequired("output-path")
	ProjectCmd.MarkFlagsOneRequired("source-set", "sources")
	ProjectCmd.MarkFlagsMutuallyExclusive("source-set", "sources")

	return ProjectCmd
}
