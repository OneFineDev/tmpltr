/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/OneFineDev/tmpltr/internal/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// Viper can access a nested field in configs by passing a . delimited path of keys. Viper lookups are case insensitive.
	rootCfgKeyVerbose          string = "verbose"
	rootCfgKeyFlagDebug        string = "flagDebug"
	rootCfgKeySourceConfigFile string = "sourceConfigFile"
	rootCfgKeyLoggingLevel     string = "logging.level"
	rootCfgKeyLoggingFormat    string = "logging.format"
	rootCfgKeyLoggingOutputs   string = "logging.outputs"
)

var (
	globalCfg                  *GlobalConfig = &GlobalConfig{}
	cfgFile                    string
	replaceHyphenWithCamelCase = true

	// Preserves the flag/config override logic in bindFlags() even with nested config keys in config file.
	flagToViperKeyLookup map[string]string = map[string]string{
		"log-level":          rootCfgKeyLoggingLevel,
		"log-format":         rootCfgKeyLoggingFormat,
		"log-output":         rootCfgKeyLoggingOutputs,
		"source-config-file": rootCfgKeySourceConfigFile,
		"verbose":            rootCfgKeyVerbose,
	}
)

var appLogger *slog.Logger

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "tmpltr",
		Short: "tmpltr is a tool for dynamically scaffolding project repositories",
		Long: `
	tmpltr allows the building of project repositories using composition from multiple "sources".
	Sources are file storage locations, with git, blob and file sources currently supported.
	Files within a source can be template files, and values for these templates can be provided via
	a values file or via the cli when you execute project creation. tmpltr can also add "partials"
	to an existing repository. You can also use tmpltr to capture a path as a local source, which can
	then be committed to a remote.
	`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// You can bind cobra and viper in a few locations, but PersistencePreRunE on the root command works well
			err := initConfig(cmd)
			appLogger = logger.InitLogger(globalCfg.Level, globalCfg.Format, os.Stdout)
			return err
		},
		// Uncomment the following line if your bare application
		// has an action associated with it:
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Usage()
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(
		&cfgFile, "config", "", "path to folder containing config file (not the path of the file itself)",
	)
	rootCmd.PersistentFlags().StringVarP(
		&globalCfg.SourceConfigFile, "source-config-file", "s", "$HOME/.tmpltr/.sources.yaml", "path to sources config file",
	)
	rootCmd.PersistentFlags().BoolVarP(
		&globalCfg.Verbose, "verbose", "v", false, "Verbose mode",
	)
	rootCmd.PersistentFlags().StringVar(
		&globalCfg.LoggingConfig.Level, "log-level", "INFO", rootCfgKeyLoggingLevel,
	)
	rootCmd.PersistentFlags().StringVar(
		&globalCfg.LoggingConfig.Format, "log-format", "text", rootCfgKeyLoggingFormat,
	)
	rootCmd.PersistentFlags().StringSliceVar(
		&globalCfg.LoggingConfig.Outputs, "log-output", []string{"StdOut"}, rootCfgKeyLoggingOutputs,
	)

	rootCmd.AddCommand(
		NewGetCommand(),
		NewCreateCommand(),
		NewProjectCommand(),
	)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cmd := NewRootCommand()
	err := cmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	if cfgFile != "" {
		viper.SetConfigName(".tmpltr")
		viper.SetConfigType("yaml")
		// Use config file from the flag.
		viper.AddConfigPath(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		tmpltrConfigPath := path.Join(home, ".tmpltr")

		// Search config in home directory with name ".tmpltr" (without extension).
		viper.AddConfigPath(tmpltrConfigPath)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tmpltr")
	}

	viper.SetEnvPrefix("TMPLTR")
	viper.SetDefault(rootCfgKeyVerbose, false)
	viper.SetDefault(rootCfgKeySourceConfigFile, "/home/parisb/repos/PLT.PRODUCT.TMPLTR/tmpltr/test/.tmpltr")
	viper.SetDefault(rootCfgKeyLoggingFormat, "text")
	viper.SetDefault(rootCfgKeyLoggingLevel, "INFO")
	viper.SetDefault(rootCfgKeyLoggingOutputs, []string{"StdOut"})

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if globalCfg.Verbose {
			fmt.Fprintln(os.Stderr, "using config file:", viper.ConfigFileUsed())
		}
	} else {
		return fmt.Errorf("config read error %s:", err.Error())
	}

	bindFlags(cmd, viper.GetViper())
	return nil
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable).
func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Determine the naming convention of the flags when represented in the config file
		configName := f.Name

		if !f.Changed && viper.IsSet(flagToViperKeyLookup[configName]) {
			val := viper.Get(flagToViperKeyLookup[configName])
			if viper.GetBool(rootCfgKeyFlagDebug) {
				fmt.Printf("flag %v default value %v is overridden by viper value %v\n", f.Name, f.DefValue, val)
			}
			_ = cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
