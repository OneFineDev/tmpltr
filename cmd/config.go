package cmd

type GlobalConfig struct {
	LoggingConfig
	Verbose          bool
	SourceConfigFile string
}

type LoggingConfig struct {
	Level   string
	Format  string
	Outputs []string
}
