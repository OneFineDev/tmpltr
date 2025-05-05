//go:build integration

package cmd_test

import (
	"io"
	"testing"

	"github.com/cucumber/godog"
	"github.com/spf13/cobra"
)

func TestValues(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		// TestSuiteInitializer: IntializeTestSuite,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/values.feature"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

type valuesScenarioContainer struct {
	ProjectCommand        *cobra.Command
	TmpltrConfigPathFlag  string
	SourcesConfigPathFlag string
	ValuesPathFlag        string
	OutputPathFlag        string
	OutputPath            string
	CommandArgs           []string
	OutPutWriter          io.Writer
	Context               *godog.ScenarioContext
}
