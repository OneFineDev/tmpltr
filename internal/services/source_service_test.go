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

func TestParseSourceConfigFile(t *testing.T) {
	// Arrange
	tests := []struct {
		name           string
		fileContent    string
		expectError    bool
		expectedConfig *types.SourceConfig
	}{
		{
			name: "Valid YAML file",
			fileContent: `
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
			expectedConfig: &types.SourceConfig{
				SourceAuths: []types.SourceAuth{
					{
						AuthAlias: "azureDevOpsENVPAT",
						UserName:  "parisbrooker@parisbrooker.co.uk",
					},
					{
						AuthAlias: "azureDevOpsPAT",
						UserName:  "parisbrooker@parisbrooker.co.uk",
						Pat:       "09876",
					},
					{
						AuthAlias: "azureDevOpsSSH",
						UserName:  "parisbrooker@parisbrooker.co.uk",
						SSHKey:    "/home/parisb/.ssh/ado",
					},
				},
				SourceSets: []types.SourceSet{
					{
						Alias:   "terraformChildSet",
						Sources: []string{"terraformChild", "doc", "vscode", "common"},
						Values: map[string]string{
							"terraformVersionConstraintString": ">= 1, < 2",
						},
					},
					{
						Alias:   "terraformDeploymentSet",
						Sources: []string{"vscode", "terraformDeployment", "common", "doc"},
						Values: map[string]string{
							"terraformVersionConstraintString": ">= 1, < 2",
							"terraformVersion":                 "1.10.5",
						},
					},
					{
						Alias:   "goWebSet",
						Sources: []string{"goWeb", "goTooling", "vscode", "common", "doc"},
					},
					{
						Alias:   "goServiceSet",
						Sources: []string{"goService", "goTooling", "vscode", "common", "doc"},
					},
				},
				Sources: []types.Source{
					{
						Alias:           "terraformChild",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.child",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "terraformDeployment",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.deployment",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "vscode",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.vscode",
						Path:            "/terraform",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "common",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.repo",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "doc",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.docs",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "goWeb",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.go.web",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "goService",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.go.service",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
					{
						Alias:           "goTooling",
						SourceType:      "git",
						URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.go.tooling",
						Path:            "/",
						SourceAuthAlias: "azureDevOpsSSH",
					},
				},
			},
		},
		{
			name:        "Invalid YAML file",
			fileContent: `invalid_yaml: [`,
			expectError: true,
		},
		{
			name:        "Empty file",
			fileContent: ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			fs := afero.NewMemMapFs()
			testFilePath := ".sources.yaml"
			err := afero.WriteFile(fs, testFilePath, []byte(tt.fileContent), 0644)
			require.NoError(t, err)

			f, err := fs.Open(testFilePath)
			require.NoError(t, err)

			// Act
			result, err := services.ParseSourceConfigFile(f)

			// Make sure we close the file after using it
			f.Close()

			// Assert
			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedConfig, result)
			}
		})
	}
}
