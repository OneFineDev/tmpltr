package services //nolint: testpackage // no export

// import (
// 	"context"
// 	"errors"
// 	"log/slog"
// 	"path"
// 	"path/filepath"
// 	"sync"
// 	"testing"
// 	"time"

// 	"github.com/OneFineDev/tmpltr/internal/storage"
// 	"github.com/OneFineDev/tmpltr/internal/types"
// 	"github.com/go-git/go-billy/v5"
// 	"github.com/go-git/go-billy/v5/memfs"
// 	"github.com/spf13/afero"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// var tesDirPath string = path.Join("..", "..", "test")

// func testSourcesCommandConfig() *SourcesCommandConfig {
// 	tmpltrConfigPath, _ := filepath.Abs(path.Join(tesDirPath, ".tmpltr", ".tmpltr"))
// 	sourceConfigFilePath, _ := filepath.Abs(path.Join(tesDirPath, ".tmpltr", ".types.yaml"))
// 	outputPath, _ := filepath.Abs(path.Join(tesDirPath, "output"))

// 	return &SourcesCommandConfig{
// 		TmpltrConfigPath:     tmpltrConfigPath,
// 		SourceConfigFilePath: sourceConfigFilePath,
// 		OutputPath:           outputPath,
// 	}
// }

// // TestBuildProjectSourceConfigs tests the BuildProjectSourceConfigs method.
// func TestBuildProjectSourceConfigs(t *testing.T) {
// 	t.Setenv("TMLPTR_azureDevOpsENVPAT_PAT", "12345")
// 	tests := []struct {
// 		name             string
// 		sourceConfigPath string
// 		expectedError    error
// 		configDoc        string
// 	}{
// 		{
// 			name:          "Valid config file",
// 			expectedError: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			ss := &SourceService{
// 				SourcesCommandConfig: testSourcesCommandConfig(),
// 			}
// 			// Act
// 			err := ss.BuildProjectSourceConfigs()

// 			// Assert
// 			if tt.expectedError != nil {
// 				require.Error(t, err)
// 				assert.EqualError(t, err, tt.expectedError.Error())
// 			} else {
// 				require.NoError(t, err)
// 				assert.NotNil(t, ss.SourceConfig)

// 				// Check we have the expected source set keys
// 				sourceSetMapKeys := getMapKeys(ss.SourceSets)
// 				assert.ElementsMatch(t, []string{
// 					"terraformChildSet",
// 					"terraformDeploymentSet",
// 					"goWebSet",
// 					"goServiceSet",
// 				}, sourceSetMapKeys)

// 				// Check we have expected source keys
// 				sourcesAliasKeys := getMapKeys(ss.SourceMap)
// 				assert.ElementsMatch(t, []string{
// 					"terraformChild",
// 					"terraformDeployment",
// 					"vscode",
// 					"common",
// 					"doc",
// 					"goWeb",
// 					"goService",
// 					"goTooling",
// 				}, sourcesAliasKeys)

// 				// Check we have the expected SourceAuths and PAT gets set by env var
// 				assert.Equal(t, "12345", ss.SourceAuthMap["azureDevOpsENVPAT"].Pat)
// 				assert.Equal(t, "09876", ss.SourceAuthMap["azureDevOpsPAT"].Pat)
// 				assert.Equal(t, "/home/parisb/.ssh/ado", ss.SourceAuthMap["azureDevOpsSSH"].SshKey)

// 				// Check we have expected clients
// 				assert.IsType(t, &storage.GitClient{}, ss.SourceClients["git"])
// 			}
// 		})
// 	}
// }

// func TestSetTargetSourcesFromSourceSet(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		alias         string
// 		sourceSets    map[string]types.SourceSet
// 		sourceMap     map[string]types.Source
// 		expectedError error
// 		expectedKeys  []string
// 	}{
// 		{
// 			name:  "Valid source set alias",
// 			alias: "terraformDeploymentSet",
// 			sourceSets: map[string]types.SourceSet{
// 				"terraformDeploymentSet": {
// 					Alias:   "terraformDeploymentSet",
// 					Sources: []string{"terraformDeployment", "common"},
// 				},
// 			},
// 			sourceMap: map[string]types.Source{
// 				"terraformDeployment": {Alias: "terraformDeployment"},
// 				"common":              {Alias: "common"},
// 			},
// 			expectedError: nil,
// 			expectedKeys:  []string{"terraformDeployment", "common"},
// 		},
// 		{
// 			name:  "Source not found in source map",
// 			alias: "terraformDeploymentSet",
// 			sourceSets: map[string]types.SourceSet{
// 				"terraformDeploymentSet": {
// 					Alias:   "terraformDeploymentSet",
// 					Sources: []string{"terraformDeployment", "missingSource"},
// 				},
// 			},
// 			sourceMap: map[string]types.Source{
// 				"terraformDeployment": {Alias: "terraformDeployment"},
// 			},
// 			expectedError: errors.New("source not found: missingSource"),
// 			expectedKeys:  nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ss := &SourceService{
// 				SourceSets:    tt.sourceSets,
// 				SourceMap:     tt.sourceMap,
// 				TargetSources: make(map[string]types.Source),
// 			}

// 			err := ss.setTargetSourcesFromSourceSet(tt.alias)

// 			if tt.expectedError != nil {
// 				require.Error(t, err)
// 				assert.EqualError(t, err, tt.expectedError.Error())
// 			} else {
// 				require.NoError(t, err)
// 				targetSourceKeys := getMapKeys(ss.TargetSources)
// 				assert.ElementsMatch(t, tt.expectedKeys, targetSourceKeys)
// 			}
// 		})
// 	}
// }

// func TestSetSourceAuthForSource(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		targetSources map[string]types.Source
// 		sourceAuthMap map[string]types.SourceAuth
// 		expectedError error
// 	}{
// 		{
// 			name: "Valid source auths",
// 			targetSources: map[string]types.Source{
// 				"source1": {Alias: "source1", SourceAuthAlias: "auth1"},
// 				"source2": {Alias: "source2", SourceAuthAlias: "auth2"},
// 			},
// 			sourceAuthMap: map[string]types.SourceAuth{
// 				"auth1": {AuthAlias: "auth1", Pat: "pat1"},
// 				"auth2": {AuthAlias: "auth2", Pat: "pat2"},
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "Missing source auth",
// 			targetSources: map[string]types.Source{
// 				"source1": {Alias: "source1", SourceAuthAlias: "auth1"},
// 				"source2": {Alias: "source2", SourceAuthAlias: "missingAuth"},
// 			},
// 			sourceAuthMap: map[string]types.SourceAuth{
// 				"auth1": {AuthAlias: "auth1", Pat: "pat1"},
// 			},
// 			expectedError: errors.New("source auth not found: missingAuth"),
// 		},
// 		{
// 			name: "No source auth alias",
// 			targetSources: map[string]types.Source{
// 				"source1": {Alias: "source1"},
// 				"source2": {Alias: "source2"},
// 			},
// 			sourceAuthMap: map[string]types.SourceAuth{},
// 			expectedError: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ss := &SourceService{
// 				TargetSources: tt.targetSources,
// 				SourceAuthMap: tt.sourceAuthMap,
// 			}

// 			err := ss.setSourceAuthForSources()

// 			if tt.expectedError != nil {
// 				require.Error(t, err)
// 				assert.EqualError(t, err, tt.expectedError.Error())
// 			} else {
// 				require.NoError(t, err)
// 				for _, source := range ss.TargetSources {
// 					if source.SourceAuthAlias != "" {
// 						assert.NotNil(t, source.SourceAuth)
// 						assert.Equal(t, tt.sourceAuthMap[source.SourceAuthAlias], *source.SourceAuth)
// 					} else {
// 						assert.Nil(t, source.SourceAuth)
// 					}
// 				}
// 			}
// 		})
// 	}
// }

// func TestCloneSources(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		targetSources  map[string]types.Source
// 		sourceClients  map[string]SourceClient
// 		expectedErrors []string
// 	}{
// 		{
// 			name: "Successful clone for all sources",
// 			targetSources: map[string]types.Source{
// 				"source1": {Alias: "source1", SourceType: types.GitSourceType},
// 				"source2": {Alias: "source2", SourceType: types.FileSourceType},
// 			},
// 			sourceClients: map[string]SourceClient{
// 				string(types.GitSourceType):  &MockSourceClient{},
// 				string(types.FileSourceType): &MockSourceClient{},
// 			},
// 			expectedErrors: nil,
// 		},
// 		{
// 			name:           "No target sources",
// 			targetSources:  map[string]types.Source{},
// 			sourceClients:  map[string]SourceClient{},
// 			expectedErrors: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			ss := &SourceService{
// 				TargetSources: tt.targetSources,
// 				SourceClients: tt.sourceClients,
// 				SourcesCommandConfig: &SourcesCommandConfig{
// 					OutputPath: "/output/path",
// 				},
// 				Logger: slog.New(slog.DiscardHandler),
// 			}
// 			targetFs := afero.NewMemMapFs()
// 			ctx := context.Background()

// 			// Act
// 			resultFs, errs := ss.CloneSources(ctx, targetFs)

// 			// Assert
// 			if tt.expectedErrors != nil {
// 				require.NotNil(t, errs)
// 				require.Len(t, errs, len(tt.expectedErrors))
// 				for i, err := range errs {
// 					assert.Contains(t, err.Error(), tt.expectedErrors[i])
// 				}
// 				assert.Nil(t, resultFs)
// 			} else {
// 				require.Nil(t, errs)
// 				assert.NotNil(t, resultFs)
// 			}
// 		})
// 	}
// }

// // MockSourceClient implements SourceClient for testing.
// type MockSourceClient struct {
// 	CurrentSource  *types.Source
// 	CloneCalledFor []string
// 	ReturnFs       billy.Filesystem
// 	ReturnError    error
// 	Delay          time.Duration
// 	mu             sync.Mutex
// }

// func (m *MockSourceClient) CloneSource(ctx context.Context, opts storage.CloneOpts) (billy.Filesystem, error) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()

// 	if m.CurrentSource != nil {
// 		m.CloneCalledFor = append(m.CloneCalledFor, m.CurrentSource.Alias)
// 	}

// 	if m.Delay > 0 {
// 		time.Sleep(m.Delay)
// 	}

// 	if m.ReturnFs == nil {
// 		m.ReturnFs = memfs.New()
// 	}

// 	return m.ReturnFs, m.ReturnError
// }

// func (m *MockSourceClient) GetCurrentSource() *types.Source {
// 	return m.CurrentSource
// }

// func (m *MockSourceClient) SetCurrentSource(s *types.Source) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()
// 	m.CurrentSource = s
// }

// type mockSourceClient struct {
// 	cloneSourceFunc func(ctx context.Context, opts storage.CloneOpts) error
// }

// func (m *mockSourceClient) CloneSource(ctx context.Context, opts storage.CloneOpts) error {
// 	return m.cloneSourceFunc(ctx, opts)
// }

// func (m *mockSourceClient) GetCurrentSource() *types.Source {
// 	return nil
// }

// func (m *mockSourceClient) SetCurrentSource(s *types.Source) {}

// func getMapKeys[K comparable, V any](m map[K]V) []K {
// 	keys := make([]K, 0, len(m))
// 	for k := range m {
// 		keys = append(keys, k)
// 	}
// 	return keys
// }
