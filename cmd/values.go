package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/OneFineDev/tmpltr/internal/storage"
	package_errors "github.com/OneFineDev/tmpltr/internal/tmpltrerrors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const tempPath string = "temp" //  since this will always run in mem, making the "output" path constant

func NewValuesCommand() *cobra.Command { //nolint:funlen
	ValuesCmd := &cobra.Command{
		Use:   "values",
		Short: "A brief description of your command",
		Long: `A longer description that spans multiple lines and likely contains examples
	and usage of using your command. For example:

	Cobra is a CLI library for Go that empowers applications.
	This application is a tool to generate the needed files
	to quickly Values a Cobra application.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			sourceConfigFile, err := os.Open(globalCfg.SourceConfigFile)
			if err != nil {
				return fmt.Errorf("error opening source config file: %w", err)
			}
			defer sourceConfigFile.Close()

			parsedSorcesConfig, err := services.ParseSourceConfigFile(sourceConfigFile)
			if err != nil {
				return fmt.Errorf("error parsing source config file: %w", err)
			}

			ss := services.NewSourceService(sourceCmdCfg, appLogger, cmd.Name())

			ss.Logger.Info("values called")

			err = ss.BuildProjectSourceConfigs(parsedSorcesConfig)
			if err != nil {
				return fmt.Errorf("error building source configs: %w", err)
			}
			ctx := context.Background()

			memFs := afero.NewMemMapFs()

			safeFs := &storage.SafeFs{
				Fs: memFs,
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
					_ = safeFs.CopyFileSystemSafe(b, "/", tempPath)
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

			// for {
			// 	select {
			// 	case r, more := <-billyChan:
			// 		if !more {
			// 		}
			// 		safeFs.CopyFileSystemSafe(r, "/", ss.OutputPath)

			// 	case e := <-errChan:
			// 		mu.Lock()
			// 		cloneErrs = append(cloneErrs, e)
			// 		mu.Unlock()
			// 	}
			// }

			ts := services.NewTemplateService(safeFs)

			// Template handling
			err = ts.GetTemplateFiles(tempPath)
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

			_, _ = out.Write(p)

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
