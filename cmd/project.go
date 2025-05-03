package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/OneFineDev/tmpltr/internal/storage"
	package_errors "github.com/OneFineDev/tmpltr/internal/tmpltrerrors"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var sourceCmdCfg *services.SourcesCommandConfig = &services.SourcesCommandConfig{} //nolint:gochecknoglobals //will fix

func NewProjectCommand() *cobra.Command { //nolint:gocognit,funlen
	ProjectCmd := &cobra.Command{
		Use:   "project",
		Short: "Builds a project from the specified SourceSet/Sources",
		Long: `Project builds a project from the specified SourceSet/Sources
which are defined in your SourcesConfig file. Where the Sources contain template
values which need to be provided, these can be provided interactively or by
passing a values file to the command on the --values-file flag. See 'get values'
command documentation for an easy way to produce values files.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()

			sourceConfigFile, err := os.Open(globalCfg.SourceConfigFile)
			if err != nil {
				return fmt.Errorf(package_errors.OpenSourceConfigFileError, err)
			}
			defer sourceConfigFile.Close()

			parsedSourcesConfig, err := services.ParseSourceConfigFile(sourceConfigFile)
			if err != nil {
				return fmt.Errorf(package_errors.ParseSourceConfigFileError, err)
			}

			ss := services.NewSourceService(sourceCmdCfg, appLogger, cmd.Name())

			err = ss.BuildProjectSourceConfigs(parsedSourcesConfig)
			if err != nil {
				return fmt.Errorf(package_errors.BuildSourceConfigError, err)
			}

			// As we're building a project this will be a real Os FS
			osFs := afero.NewOsFs()

			// The output path may or may not exist
			_, err = osFs.Stat(sourceCmdCfg.OutputPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// Create the directory if it doesn't exist
					err = osFs.MkdirAll(sourceCmdCfg.OutputPath, 0755) //nolint:mnd
					if err != nil {
						return fmt.Errorf("failed to create output directory: %w", err)
					}
				} else {
					// Handle other potential errors
					return fmt.Errorf("failed to check output path: %w", err)
				}
			}

			safeFs := &storage.SafeFs{
				Fs: osFs,
			}

			mu := sync.Mutex{}

			var receivedErrors []error

			// Clones the sources concurrently
			billyChan, errChan := ss.CloneSources(ctx)

			// Write to target fs sequentially

			var wg sync.WaitGroup

			wg.Add(2) //nolint:mnd
			go func() {
				defer wg.Done()
				for b := range billyChan {
					_ = safeFs.CopyFileSystemSafe(b, "/", sourceCmdCfg.OutputPath)
				}
			}()

			go func() {
				defer wg.Done()
				for e := range errChan {
					mu.Lock()
					receivedErrors = append(receivedErrors, e)
					mu.Unlock()
				}
			}()

			wg.Wait()

			if len(receivedErrors) > 0 {
				return package_errors.FlattenCloneErrors(receivedErrors...)
			}

			ts := services.NewTemplateService(safeFs)

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

			if sourceCmdCfg.ValuesFilePath == "" {
				e := ts.InteractiveInput()
				if e != nil {
					return e
				}
			} else {
				f, e := os.Open(sourceCmdCfg.ValuesFilePath)
				if e != nil {
					return fmt.Errorf(package_errors.OpenValuesFileError, e)
				}

				ts.TemplateValuesMap, e = services.ReadYamlFromFile[types.TemplateValuesMap](f)
				if e != nil {
					return fmt.Errorf(package_errors.OpenValuesFileError, e)
				}
			}

			err = ts.ExecuteTemplates()
			if err != nil {
				return fmt.Errorf(package_errors.TemplateExecutionError, err)
			}

			err = ts.RenameTargetTemplateFiles()
			if err != nil {
				return fmt.Errorf(package_errors.TemplateFileRenameError, err)
			}
			return nil
		},
	}

	ProjectCmd.Flags().StringVarP(
		&sourceCmdCfg.OutputPath, "output-path", "o", "", "the path in which the SourceSet/Sources will be rendered",
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
	ProjectCmd.Flags().StringVarP(
		&sourceCmdCfg.ValuesFilePath, "values-file", "f", "", "path to a values file used to populate template values. falls into interactive mode if not provided.",
	)
	ProjectCmd.Flags().BoolVarP(
		&sourceCmdCfg.FailOnMissingTemplateValue, "fail-on-missing-value", "m", false, "whether to fail project generation is a template input value is missing.",
	)

	_ = ProjectCmd.MarkFlagRequired("output-path")
	ProjectCmd.MarkFlagsOneRequired("source-set", "sources")
	ProjectCmd.MarkFlagsMutuallyExclusive("source-set", "sources")

	return ProjectCmd
}
