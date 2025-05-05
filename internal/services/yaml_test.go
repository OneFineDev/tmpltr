//go:build !integration

package services_test

import (
	"testing"

	"github.com/OneFineDev/tmpltr/internal/services"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadYamlFromFile(t *testing.T) { //nolint:gocognit
	tests := []struct {
		name           string
		yamlContent    string
		expectError    bool
		expectedValues struct {
			sourceAuthCount  int
			sourceSetCount   int
			sourceCount      int
			sourceSetsToTest map[string][]string
		}
	}{
		{
			name: "Valid sources YAML document",
			yamlContent: `# yaml-language-server: $schema=./sources.schema.json
sourceAuths:
  - authAlias: "azureDevOpsENVPAT"
    userName: "parisbrooker@parisbrooker.co.uk"

  - authAlias: "azureDevOpsPAT"
    userName: "parisbrooker@parisbrooker.co.uk"
    pat: "09876"

  - authAlias: "azureDevOpsSSH"
    userName: "parisbrooker@parisbrooker.co.uk"
    sshKeyPath: "/home/parisb/.ssh/ado" # If present, TMLPTR_DEFAULT_SSH_KEY_PATH environment variable will overwrite this value, or will be used if this value is not present

sourceSets:
  - alias: terraformChildSet
    sources:
      - terraformChild
      - doc
      - vscode
      - common
    values:
      terraformVersionConstraintString: ">= 1, < 2"
  - alias: terraformDeploymentSet
    sources:
      - vscode
      - terraformDeployment
      - common
      - doc
    values:
      terraformVersionConstraintString: ">= 1, < 2"
      terraformVersion: "1.10.5"
  - alias: goWebSet
    sources:
      - goWeb
      - goTooling
      - vscode
      - common
      - doc
  - alias: goServiceSet
    sources:
      - goService
      - goTooling
      - vscode
      - common
      - doc

sources:
  - alias: terraformChild
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.child"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: terraformDeployment
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.deployment"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: vscode
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.vscode"
    path: "/terraform"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: common
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.repo"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: doc
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.docs"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: goWeb
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.go.web"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: goService
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.go.service"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"

  - alias: goTooling
    sourceType: git
    url: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.go.tooling"
    path: "/"
    sourceAuthAlias: "azureDevOpsSSH"
`,
			expectError: false,
			expectedValues: struct {
				sourceAuthCount  int
				sourceSetCount   int
				sourceCount      int
				sourceSetsToTest map[string][]string
			}{
				sourceAuthCount: 3,
				sourceSetCount:  4,
				sourceCount:     8,
				sourceSetsToTest: map[string][]string{
					"terraformChildSet": {"terraformChild", "doc", "vscode", "common"},
					"goServiceSet":      {"goService", "goTooling", "vscode", "common", "doc"},
				},
			},
		},
		{
			name:        "Invalid YAML",
			yamlContent: `: invalid yaml - this is not a valid yaml document`,
			expectError: true,
			expectedValues: struct {
				sourceAuthCount  int
				sourceSetCount   int
				sourceCount      int
				sourceSetsToTest map[string][]string
			}{},
		},
		{
			name:        "Empty YAML",
			yamlContent: ``,
			expectError: true,
			expectedValues: struct {
				sourceAuthCount  int
				sourceSetCount   int
				sourceCount      int
				sourceSetsToTest map[string][]string
			}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			testFilePath := "/test/.sources.yaml"
			err := afero.WriteFile(fs, testFilePath, []byte(tt.yamlContent), 0644)
			require.NoError(t, err)
			f, err := fs.Open(testFilePath)
			require.NoError(t, err)
			// Act
			result, err := services.ReadYamlFromFile[types.SourceConfig](f)
			f.Close()
			// Assert
			if tt.expectError {
				assert.Error(t, err, "Expected an error but got none")
			} else {
				require.NoError(t, err, "Did not expect an error but got: %v", err)

				// Verify source auth count
				assert.Equal(t, tt.expectedValues.sourceAuthCount, len(result.SourceAuths), //nolint:testifylint //no prob
					"Expected %d source auths, but got %d", tt.expectedValues.sourceAuthCount, len(result.SourceAuths))

				// Verify source set count
				assert.Equal(t, tt.expectedValues.sourceSetCount, len(result.SourceSets), //nolint:testifylint //no prob
					"Expected %d source sets, but got %d", tt.expectedValues.sourceSetCount, len(result.SourceSets))

				// Verify source count
				assert.Equal(t, tt.expectedValues.sourceCount, len(result.Sources), //nolint:testifylint //no prob
					"Expected %d sources, but got %d", tt.expectedValues.sourceCount, len(result.Sources))

				// Verify specific source sets contents
				for setName, expectedSources := range tt.expectedValues.sourceSetsToTest {
					var sourceSet *types.SourceSet
					for i := range result.SourceSets {
						if result.SourceSets[i].Alias == setName {
							sourceSet = &result.SourceSets[i]
							break
						}
					}

					require.NotNil(t, sourceSet, "Source set %s not found", setName)
					assert.ElementsMatch(t, expectedSources, sourceSet.Sources,
						"Sources in %s do not match expected values", setName)
				}

				// Verify specific sources (sample check)
				foundTerraformChild := false
				foundGoService := false

				for _, source := range result.Sources {
					if source.Alias == "terraformChild" {
						foundTerraformChild = true
						assert.Equal(t, types.GitSourceType, source.SourceType)
						assert.Equal(t, "/", source.Path)
						assert.Equal(t, "azureDevOpsSSH", source.SourceAuthAlias)
					}
					if source.Alias == "goService" {
						foundGoService = true
						assert.Equal(t, types.GitSourceType, source.SourceType)
						assert.Equal(t, "/", source.Path)
						assert.Equal(t, "azureDevOpsSSH", source.SourceAuthAlias)
					}
				}

				assert.True(t, foundTerraformChild, "terraformChild source not found")
				assert.True(t, foundGoService, "goService source not found")
			}
		})
	}
}
