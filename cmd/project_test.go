//go:build integration

package cmd_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/OneFineDev/tmpltr/cmd"
	"github.com/cucumber/godog"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

const (
	// testTmpltrConfigFile
	terraformChildSetValues = `
provider:
	azureApi:
		versionConstraint: ">= 2.0.0"
	azurerm:
		versionConstraint: ">= 4, <5"
	terraform:
		version: "1.11.0"
		versionConstraint: ">1, <2"
`
)

func TestProject(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		// TestSuiteInitializer: IntializeTestSuite,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"features/project.feature"},
			TestingT: t, // Testing instance that will run subtests.
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

type projectScenarioContainer struct {
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

func (psc *projectScenarioContainer) aSourceSetOfTypeAndAnAuthenticationMethodOf(arg1, arg2 string) error {

	psc.CommandArgs = []string{
		"create",
		"project",
		"--source-set=terraformChildSet",
	}
	psc.CommandArgs = append(psc.CommandArgs, psc.TmpltrConfigPathFlag, psc.ValuesPathFlag, psc.OutputPathFlag)
	return nil
}

func (psc *projectScenarioContainer) aUserRunsTheCreateProjectCommandWithTheFlagsAndTheProvided(
	arg1, arg2 string,
) error {
	psc.ProjectCommand.SetArgs(psc.CommandArgs)
	psc.ProjectCommand.SetOut(psc.OutPutWriter)
	err := psc.ProjectCommand.Execute()
	if err != nil {
		return err
	}
	return nil
}

func (psc *projectScenarioContainer) desiredFilesAreCopiedToTargetPath() error {

	t := godog.T(psc.ProjectCommand.Context())

	versionsFilePath := path.Join(psc.OutputPath, "versions.tf")

	assert.FileExists(t, versionsFilePath)

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	tmpltrCfg, _ := filepath.Abs(path.Join("..", "test", ".tmpltr"))
	// srcCfg, _ := filepath.Abs(path.Join("..", "test", ".tmpltr", ".sources.yaml"))
	outPath, _ := filepath.Abs(path.Join("..", "test", "integration", "output"))
	valuesPath, _ := filepath.Abs(path.Join("..", "test", "integration", "values.yaml"))

	tmpltrCfgFlag := fmt.Sprintf("--config=%s", tmpltrCfg)
	outPathFlag := fmt.Sprintf("--output-path=%s", outPath)
	valuesPathFlag := fmt.Sprintf("--values-file=%s", valuesPath)

	pc := &projectScenarioContainer{
		TmpltrConfigPathFlag: tmpltrCfgFlag,
		OutputPathFlag:       outPathFlag,
		OutputPath:           outPath,
		ValuesPathFlag:       valuesPathFlag,
		OutPutWriter:         bytes.NewBuffer([]byte{}),
		Context:              ctx,
	}

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		fmt.Println("creating command for scenario")
		pc.ProjectCommand = cmd.NewRootCommand()
		return ctx, nil
	})
	ctx.Step(
		`^a sourceSet of type "([^"]*)" and an authentication method of "([^"]*)"$`,
		pc.aSourceSetOfTypeAndAnAuthenticationMethodOf,
	)
	ctx.Step(
		`^a user runs the create project command with the flags "([^"]*)" and the provided "([^"]*)"$`,
		pc.aUserRunsTheCreateProjectCommandWithTheFlagsAndTheProvided,
	)
	ctx.Step(
		`^desired files are copied to target path$`, pc.desiredFilesAreCopiedToTargetPath,
	)
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		err = os.RemoveAll(pc.OutputPath)
		return ctx, err
	})
}
