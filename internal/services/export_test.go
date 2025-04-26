//go:build !integration

package services

import (
	"os"
	"testing"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestParseSourceSets(t *testing.T) {
	// Arrange
	testCases := []struct {
		name               string
		sourceSets         []types.SourceSet
		expectedSourceSets map[string]types.SourceSet
	}{
		{
			name: "Basic source sets",
			sourceSets: []types.SourceSet{
				{Alias: "set1", Sources: []string{"source1", "source2"}},
				{Alias: "set2", Sources: []string{"source3"}},
			},
			expectedSourceSets: map[string]types.SourceSet{
				"set1": {Alias: "set1", Sources: []string{"source1", "source2"}},
				"set2": {Alias: "set2", Sources: []string{"source3"}},
			},
		},
		{
			name: "Source sets with values",
			sourceSets: []types.SourceSet{
				{
					Alias:   "backend",
					Sources: []string{"api", "database"},
					Values: map[string]string{
						"framework": "express",
						"database":  "postgres",
					},
				},
				{
					Alias:   "frontend",
					Sources: []string{"web", "ui-components"},
					Values: map[string]string{
						"framework": "react",
						"styling":   "tailwind",
					},
				},
			},
			expectedSourceSets: map[string]types.SourceSet{
				"backend": {
					Alias:   "backend",
					Sources: []string{"api", "database"},
					Values: map[string]string{
						"framework": "express",
						"database":  "postgres",
					},
				},
				"frontend": {
					Alias:   "frontend",
					Sources: []string{"web", "ui-components"},
					Values: map[string]string{
						"framework": "react",
						"styling":   "tailwind",
					},
				},
			},
		},
		{
			name:               "Empty source sets",
			sourceSets:         []types.SourceSet{},
			expectedSourceSets: map[string]types.SourceSet{},
		},
		{
			name: "Source sets with empty values",
			sourceSets: []types.SourceSet{
				{
					Alias:   "microservice",
					Sources: []string{"service1", "service2"},
					Values:  map[string]string{},
				},
			},
			expectedSourceSets: map[string]types.SourceSet{
				"microservice": {
					Alias:   "microservice",
					Sources: []string{"service1", "service2"},
					Values:  map[string]string{},
				},
			},
		},
		{
			name: "Source sets with nil values",
			sourceSets: []types.SourceSet{
				{
					Alias:   "template",
					Sources: []string{"base-template"},
					Values:  nil,
				},
			},
			expectedSourceSets: map[string]types.SourceSet{
				"template": {
					Alias:   "template",
					Sources: []string{"base-template"},
					Values:  nil,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			sourceConfig := &types.SourceConfig{
				SourceSets: tc.sourceSets,
			}
			sourceService := &SourceService{
				SourceConfig: sourceConfig,
			}

			// Act
			sourceService.parseSourceSets()

			// Assert
			assert.Equal(t, len(tc.expectedSourceSets), len(sourceService.SourceSets),
				"SourceSets should have the expected number of entries")

			for alias, expectedSet := range tc.expectedSourceSets {
				actualSet, exists := sourceService.SourceSets[alias]
				assert.True(t, exists, "SourceSets should contain an entry for %s", alias)

				assert.Equal(t, expectedSet.Alias, actualSet.Alias,
					"Alias should match for %s", alias)
				assert.Equal(t, expectedSet.Sources, actualSet.Sources,
					"Sources should match for %s", alias)
				assert.Equal(t, expectedSet.Values, actualSet.Values,
					"Values should match for %s", alias)
			}
		})
	}
}

func TestParseSources(t *testing.T) { //nolint:gocognit
	// Arrange
	testCases := []struct {
		name            string
		sources         []types.Source
		expectedSources map[string]types.Source
	}{
		{
			name: "Git sources with different properties",
			sources: []types.Source{
				{
					Alias:           "terraformChild",
					SourceType:      types.GitSourceType,
					URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.child",
					Path:            "/",
					SourceAuthAlias: "azureDevOpsSSH",
				},
				{
					Alias:           "vscode",
					SourceType:      types.GitSourceType,
					URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.vscode",
					Path:            "/terraform",
					SourceAuthAlias: "azureDevOpsSSH",
				},
				{
					Alias:           "publicRepo",
					SourceType:      types.GitSourceType,
					URL:             "https://github.com/public/repository.git",
					Path:            "/",
					SourceAuthAlias: "",
				},
			},
			expectedSources: map[string]types.Source{
				"terraformChild": {
					Alias:           "terraformChild",
					SourceType:      types.GitSourceType,
					URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.child",
					Path:            "/",
					SourceAuthAlias: "azureDevOpsSSH",
				},
				"vscode": {
					Alias:           "vscode",
					SourceType:      types.GitSourceType,
					URL:             "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.common.vscode",
					Path:            "/terraform",
					SourceAuthAlias: "azureDevOpsSSH",
				},
				"publicRepo": {
					Alias:           "publicRepo",
					SourceType:      types.GitSourceType,
					URL:             "https://github.com/public/repository.git",
					Path:            "/",
					SourceAuthAlias: "",
				},
			},
		},
		{
			name: "File sources with different properties",
			sources: []types.Source{
				{
					Alias:      "localTemplate",
					SourceType: types.FileSourceType,
					Path:       "/home/user/templates/local-template",
				},
				{
					Alias:      "sharedTemplate",
					SourceType: types.FileSourceType,
					Path:       "/opt/shared/templates/common",
				},
			},
			expectedSources: map[string]types.Source{
				"localTemplate": {
					Alias:      "localTemplate",
					SourceType: types.FileSourceType,
					Path:       "/home/user/templates/local-template",
				},
				"sharedTemplate": {
					Alias:      "sharedTemplate",
					SourceType: types.FileSourceType,
					Path:       "/opt/shared/templates/common",
				},
			},
		},
		{
			name: "Mixed source types with all fields",
			sources: []types.Source{
				{
					Alias:           "fullGitRepo",
					SourceType:      types.GitSourceType,
					URL:             "git@github.com:user/repo.git",
					Path:            "/subfolder",
					SourceAuthAlias: "githubAuth",
				},
				{
					Alias:      "fullFilePath",
					SourceType: types.FileSourceType,
					Path:       "/var/templates/special",
				},
			},
			expectedSources: map[string]types.Source{
				"fullGitRepo": {
					Alias:           "fullGitRepo",
					SourceType:      types.GitSourceType,
					URL:             "git@github.com:user/repo.git",
					Path:            "/subfolder",
					SourceAuthAlias: "githubAuth",
				},
				"fullFilePath": {
					Alias:      "fullFilePath",
					SourceType: types.FileSourceType,
					Path:       "/var/templates/special",
				},
			},
		},
		{
			name:            "Empty sources list",
			sources:         []types.Source{},
			expectedSources: map[string]types.Source{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			sourceConfig := &types.SourceConfig{
				Sources: tc.sources,
			}
			sourceService := &SourceService{
				SourceConfig:  sourceConfig,
				SourceClients: make(map[string]SourceClient),
			}

			// Act
			sourceService.parseSources()

			// Assert
			assert.NotNil(t, sourceService.TargetSources, "TargetSources map should be initialized")
			assert.NotNil(t, sourceService.SourceMap, "SourceMap map should be initialized")
			assert.Equal(t, len(tc.expectedSources), len(sourceService.SourceMap),
				"SourceMap should contain expected number of sources")

			for alias, expectedSource := range tc.expectedSources {
				actualSource, exists := sourceService.SourceMap[alias]
				assert.True(t, exists, "SourceMap should contain an entry for %s", alias)

				assert.Equal(t, expectedSource.Alias, actualSource.Alias,
					"Alias should match for %s", alias)
				assert.Equal(t, expectedSource.SourceType, actualSource.SourceType,
					"SourceType should match for %s", alias)
				assert.Equal(t, expectedSource.Path, actualSource.Path,
					"Path should match for %s", alias)
				assert.Equal(t, expectedSource.URL, actualSource.URL,
					"URL should match for %s", alias)
				assert.Equal(t, expectedSource.SourceAuthAlias, actualSource.SourceAuthAlias,
					"SourceAuthAlias should match for %s", alias)
			}

			// Assert that source clients are properly initialized
			if len(tc.sources) > 0 {
				// We don't actually create clients in the parseSources function, so they should be nil
				gitSourcePresent := false
				fileSourcePresent := false

				for _, source := range tc.sources {
					if source.SourceType == types.GitSourceType {
						gitSourcePresent = true
					} else if source.SourceType == types.FileSourceType {
						fileSourcePresent = true
					}
				}

				if gitSourcePresent {
					_, exists := sourceService.SourceClients[string(types.GitSourceType)]
					assert.True(t, exists, "SourceClients should have an entry for GitSourceType")
					assert.Nil(t, sourceService.SourceClients[string(types.GitSourceType)],
						"SourceClient for GitSourceType should be nil")
				}

				if fileSourcePresent {
					_, exists := sourceService.SourceClients[string(types.FileSourceType)]
					assert.True(t, exists, "SourceClients should have an entry for FileSourceType")
					assert.Nil(t, sourceService.SourceClients[string(types.FileSourceType)],
						"SourceClient for FileSourceType should be nil")
				}
			}
		})
	}
}

func TestParseSourceAuths(t *testing.T) {
	// Arrange
	testCases := []struct {
		name                string
		sourceAuths         []types.SourceAuth
		environmentVars     map[string]string
		expectedSourceAuths map[string]types.SourceAuth
	}{
		{
			name: "Basic source auths with environment variables",
			sourceAuths: []types.SourceAuth{
				{AuthAlias: "GITHUB", UserName: "github-user", SshKey: "/path/to/key.pem"},
				{AuthAlias: "GITLAB", UserName: "gitlab-user", Key: "some-key-value"},
			},
			environmentVars: map[string]string{
				"TMLPTR_GITHUB_PAT": "github-token",
				"TMLPTR_GITLAB_PAT": "gitlab-token",
			},
			expectedSourceAuths: map[string]types.SourceAuth{
				"GITHUB": {
					AuthAlias: "GITHUB",
					UserName:  "github-user",
					Pat:       "github-token",
					SshKey:    "/path/to/key.pem",
				},
				"GITLAB": {AuthAlias: "GITLAB", UserName: "gitlab-user", Pat: "gitlab-token", Key: "some-key-value"},
			},
		},
		{
			name: "Source auths with no environment variables",
			sourceAuths: []types.SourceAuth{
				{AuthAlias: "GITHUB", UserName: "github-user", Token: "existing-token"},
				{AuthAlias: "GITLAB", UserName: "gitlab-user", Pat: "existing-pat"},
			},
			environmentVars: map[string]string{},
			expectedSourceAuths: map[string]types.SourceAuth{
				"GITHUB": {AuthAlias: "GITHUB", UserName: "github-user", Token: "existing-token"},
				"GITLAB": {AuthAlias: "GITLAB", UserName: "gitlab-user", Pat: "existing-pat"},
			},
		},
		{
			name: "Source auths with some environment variables",
			sourceAuths: []types.SourceAuth{
				{AuthAlias: "GITHUB", UserName: "github-user"},
				{AuthAlias: "AZURE_DEVOPS", UserName: "ado-user"},
				{AuthAlias: "GITLAB", UserName: "gitlab-user"},
			},
			environmentVars: map[string]string{
				"TMLPTR_GITHUB_PAT":       "github-token",
				"TMLPTR_AZURE_DEVOPS_PAT": "ado-token",
				// No environment variable for GITLAB
			},
			expectedSourceAuths: map[string]types.SourceAuth{
				"GITHUB":       {AuthAlias: "GITHUB", UserName: "github-user", Pat: "github-token"},
				"AZURE_DEVOPS": {AuthAlias: "AZURE_DEVOPS", UserName: "ado-user", Pat: "ado-token"},
				"GITLAB":       {AuthAlias: "GITLAB", UserName: "gitlab-user"},
			},
		},
		{
			name: "Source auths with different fields populated",
			sourceAuths: []types.SourceAuth{
				{AuthAlias: "GITHUB", UserName: "github-user", SshKey: "/path/to/github.key"},
				{AuthAlias: "GITLAB", UserName: "gitlab-user", Key: "gitlab-ssh-key"},
				{AuthAlias: "BITBUCKET", UserName: "bb-user", Token: "bb-token"},
				{AuthAlias: "EMPTY", UserName: "empty-user"},
			},
			environmentVars: map[string]string{
				"TMLPTR_GITHUB_PAT":    "github-override-token",
				"TMLPTR_BITBUCKET_PAT": "bitbucket-override-token",
				// No environment vars for GITLAB and EMPTY
			},
			expectedSourceAuths: map[string]types.SourceAuth{
				"GITHUB": {
					AuthAlias: "GITHUB",
					UserName:  "github-user",
					Pat:       "github-override-token",
					SshKey:    "/path/to/github.key",
				},
				"GITLAB": {AuthAlias: "GITLAB", UserName: "gitlab-user", Key: "gitlab-ssh-key"},
				"BITBUCKET": {
					AuthAlias: "BITBUCKET",
					UserName:  "bb-user",
					Token:     "bb-token",
					Pat:       "bitbucket-override-token",
				},
				"EMPTY": {AuthAlias: "EMPTY", UserName: "empty-user"},
			},
		},
		{
			name:        "Empty source auths list",
			sourceAuths: []types.SourceAuth{},
			environmentVars: map[string]string{
				"TMLPTR_GITHUB_PAT": "github-token",
			},
			expectedSourceAuths: map[string]types.SourceAuth{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			sourceConfig := &types.SourceConfig{
				SourceAuths: tc.sourceAuths,
			}

			sourceService := &SourceService{
				SourceConfig: sourceConfig,
			}

			// Set environment variables for the test
			for key, value := range tc.environmentVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Act
			sourceService.parseSourceAuths()

			// Assert
			assert.Equal(t, len(tc.expectedSourceAuths), len(sourceService.SourceAuthMap),
				"SourceAuthMap should have the expected number of entries")

			for authAlias, expectedAuth := range tc.expectedSourceAuths {
				actualAuth, exists := sourceService.SourceAuthMap[authAlias]
				assert.True(t, exists, "SourceAuthMap should contain an entry for %s", authAlias)

				assert.Equal(t, expectedAuth.AuthAlias, actualAuth.AuthAlias,
					"AuthAlias should match for %s", authAlias)
				assert.Equal(t, expectedAuth.UserName, actualAuth.UserName,
					"UserName should match for %s", authAlias)
				assert.Equal(t, expectedAuth.Pat, actualAuth.Pat,
					"Pat should match for %s", authAlias)
				assert.Equal(t, expectedAuth.SshKey, actualAuth.SshKey,
					"SshKey should match for %s", authAlias)
				assert.Equal(t, expectedAuth.Key, actualAuth.Key,
					"Key should match for %s", authAlias)
				assert.Equal(t, expectedAuth.Token, actualAuth.Token,
					"Token should match for %s", authAlias)
			}
		})
	}
}
