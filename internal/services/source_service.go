package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"sync"

	package_errors "github.com/OneFineDev/tmpltr/internal/errors"
	"github.com/OneFineDev/tmpltr/internal/storage"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
	"github.com/spf13/afero"
)

type (
	CtxKeyLogger struct{}
)

const (
	logMsgGitClone = "git_clone"
	logKeyGitRepo  = "repo"
	logKeyErr      = "error"
)

// SourcesCommandConfig represents the relevant configuration settings for any command leveraging types.
type SourcesCommandConfig struct {
	// Path to tmpltr config file
	TmpltrConfigPath string

	// Path to SourceConfigFile
	SourceConfigFilePath string

	// Path in which all Source content will be rendered, i.e. the project path
	OutputPath string

	// ProjectName
	ProjectName string

	// Append Project Name
	AppendProjectName bool

	// Sources, defined in SourceConfigFile, to be rendered in a given execution
	Sources []string

	// SourceSet, defined in SourceConfigFile, to be rendered in a given execution
	SourceSet string

	// Values file path, contents of which are used to populate template values
	ValuesFilePath string

	// Whether to throw error if template execution detects a missing template input value
	FailOnMissingTemplateValue bool
}

type SourceClient interface {
	CloneSource(ctx context.Context, cloneOpts storage.CloneOpts) (billy.Filesystem, error)
	GetCurrentSource() *types.Source
	SetCurrentSource(s *types.Source)
}

/*
SourceService is responsible for parsing the contents of SourceConfig into Source structs,
and coordinating the client initialization and fetching of the Source content as required in
an execution of a command that uses Sources.
*/
type SourceService struct {
	*SourcesCommandConfig

	Logger *slog.Logger

	SourceConfig  *types.SourceConfig
	SourceSets    map[string]types.SourceSet
	SourceMap     map[string]types.Source
	SourceAuthMap map[string]types.SourceAuth
	TargetSources map[string]types.Source
	SourceToPath  map[string][]string
	SourceClients map[string]SourceClient
}

func NewSourceService(sourcesCommandConfig *SourcesCommandConfig, logger *slog.Logger, cmdName string) *SourceService {
	if sourcesCommandConfig.ProjectName != "" && sourcesCommandConfig.AppendProjectName {
		sourcesCommandConfig.OutputPath = path.Join(sourcesCommandConfig.OutputPath, sourcesCommandConfig.ProjectName)
	}

	cmdLogger := logger.With(
		"cmd",
		cmdName,
	)

	return &SourceService{
		SourcesCommandConfig: sourcesCommandConfig,
		Logger:               cmdLogger,
	}
}

func ParseSourceConfigFile(file afero.File) (*types.SourceConfig, error) {
	srcConfig, err := ReadYamlFromFile[types.SourceConfig](file)
	if err != nil {
		wrapped := fmt.Errorf("failed to decode source config: %w", err)
		return nil, wrapped
	}
	return &srcConfig, nil
}

func (ss *SourceService) BuildProjectSourceConfigs(srcConfig *types.SourceConfig) error {
	// create the Source client map so we can add keys in parseSources()
	ss.Logger.Info("building source configs...")
	ss.SourceClients = make(map[string]SourceClient)

	ss.SourceConfig = srcConfig
	ss.parseSourceSets()
	ss.parseSources()
	ss.parseSourceAuths()
	err := ss.setTargetSourcesFromSourceSet(ss.SourceSet)
	if err != nil {
		return err
	}
	err = ss.setSourceAuthForSources()
	if err != nil {
		return err
	}
	return nil
}

func (ss *SourceService) CloneSources(ctx context.Context) (chan billy.Filesystem, chan error) {
	billyChan := make(chan billy.Filesystem, len(ss.TargetSources))
	errChan := make(chan error, len(ss.TargetSources))

	var wg sync.WaitGroup

	for _, source := range ss.TargetSources {
		wg.Add(1)
		go func(source types.Source) {
			defer wg.Done()

			ss.Logger.Info(
				logMsgGitClone, logKeyGitRepo, source.Alias,
			)
			source.Client.SetSource(&source)

			bfs, err := source.Client.Clone(ctx)
			if err != nil {
				e := &package_errors.SourceError{
					Message: fmt.Sprintf("inmem clone failed for %v", source.Alias),
					Err:     err,
				}
				ss.Logger.Error(logMsgGitClone, logKeyGitRepo, source.Alias, logKeyErr, e.Err.Error())
				errChan <- e
				return
			}

			billyChan <- bfs
		}(source)
	}

	go func() {
		wg.Wait()
		close(billyChan)
		close(errChan)
	}()

	return billyChan, errChan

	// 	// for {
	// 	// 	select {
	// 	// 	case r, more := <-billyChan:
	// 	// 		if !more {
	// 	// 			return errs
	// 	// 		}
	// 	// 		sfs.CopyFileSystemSafe(r, "/", ss.OutputPath)

	// // 	case e := <-errChan:
	// // 		mu.Lock()
	// // 		errs = append(errs, e)
	// // 		mu.Unlock()
	// // 	}
	// // }
}

func (s *SourceService) parseSourceSets() {
	s.SourceSets = make(map[string]types.SourceSet)
	for _, sourceSet := range s.SourceConfig.SourceSets {
		s.SourceSets[sourceSet.Alias] = sourceSet
	}
}

func (s *SourceService) parseSources() {
	s.TargetSources = make(map[string]types.Source)
	s.SourceMap = make(map[string]types.Source)
	for _, source := range s.SourceConfig.Sources {
		s.SourceClients[string(source.SourceType)] = nil
		s.SourceMap[source.Alias] = source
	}
}

// parseSourceAuths initializes the SourceAuthMap by iterating over the SourceAuths
// defined in the SourceConfig. For each SourceAuth, it attempts to retrieve a
// Personal Access Token (PAT) from the environment variables using a key formatted
// as "TMLPTR_<AuthAlias>_PAT". If a PAT is found, it updates the corresponding
// SourceAuth in the SourceAuthMap with the retrieved PAT.
func (s *SourceService) parseSourceAuths() {
	s.SourceAuthMap = make(map[string]types.SourceAuth)
	for _, sourceAuth := range s.SourceConfig.SourceAuths {
		s.SourceAuthMap[sourceAuth.AuthAlias] = sourceAuth

		envVarString := fmt.Sprintf("TMLPTR_%s_PAT", sourceAuth.AuthAlias)

		pat := os.Getenv(envVarString)
		if pat != "" {
			auth := s.SourceAuthMap[sourceAuth.AuthAlias]
			auth.Pat = pat
			s.SourceAuthMap[sourceAuth.AuthAlias] = auth
		}
	}
}

// createSourceClients initializes the SourceClients map with appropriate client
// instances based on the source type. It iterates over the keys of the SourceClients
// map and assigns a new client instance for each supported source type:
// - GitSourceType: Initializes a Git client using storage.NewGitClient().
// - FileSourceType: Initializes a File client using storage.NewFileClient().
// - BlobSourceType: Initializes a Blob client using storage.NewBlobClient().
func createSourceClients(t types.SourceType) (types.SourceCloner, error) {
	switch t {
	case types.GitSourceType:
		return storage.NewGitClient(), nil
	case types.FileSourceType:
		return storage.NewFileClient(), nil
	case types.BlobSourceType:
		return storage.NewBlobClient(), nil
	default:
		return nil, fmt.Errorf("failed to create client for source")
	}
}

// setTargetSources sets the target sources for a given alias by iterating through
// the source aliases in the SourceSets map. It retrieves each source from the
// SourceMap and adds it to the TargetSources map. It also inits the source client
// on the source If a source alias is not found in the SourceMap, an error is returned.
//
// Parameters:
//   - alias: The key used to identify the set of sources in the SourceSets map.
//
// Returns:
//   - error: An error is returned if a source alias is not found in the SourceMap.
func (s *SourceService) setTargetSourcesFromSourceSet(alias string) error {
	for _, sourceAlias := range s.SourceSets[alias].Sources {
		source, ok := s.SourceMap[sourceAlias]
		if !ok {
			return fmt.Errorf("source not found: %s", sourceAlias)
		}
		var err error
		// Now we know we'll be using this source, initialize its client
		source.Client, err = createSourceClients(source.SourceType)
		if err != nil {
			return fmt.Errorf("%s: %w", sourceAlias, err)
		}
		s.TargetSources[sourceAlias] = source
	}
	return nil
}

func (s *SourceService) setSourceAuthForSources() error {
	for _, source := range s.TargetSources {
		if source.SourceAuthAlias != "" {
			sourceAuth, ok := s.SourceAuthMap[source.SourceAuthAlias]
			if !ok {
				return fmt.Errorf("source auth not found: %s", source.SourceAuthAlias)
			}
			source.SourceAuth = &sourceAuth
			s.TargetSources[source.Alias] = source
		}
	}
	return nil
}
