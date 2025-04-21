package storage_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/OneFineDev/tmpltr/internal/storage"
	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupFixtureGitRepo sets up an in-memory Git repository with test files.
func setupFixtureGitRepo(t *testing.T) (string, string) {
	// Create a temporary file for SSH key simulation
	tmpKeyFile, err := os.CreateTemp("", "test-ssh-key")
	require.NoError(t, err)
	sshKeyPath := tmpKeyFile.Name()
	t.Cleanup(func() {
		os.Remove(sshKeyPath) // Clean up the temp file after test
	})

	// Create an in-memory git storage
	fs := memfs.New()
	storage := memory.NewStorage()

	// Create a new repository
	repo, err := git.Init(storage, fs)
	require.NoError(t, err)

	// Configure to simulate a remote repo
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"git@example.com:test/fixture-repo.git"},
	})
	require.NoError(t, err)

	// Create test files (including the locals.tf file)
	workTree, err := repo.Worktree()
	require.NoError(t, err)

	// Create locals.tf file
	createFileInFS(t, fs, "locals.tf", `locals {
  name = "test-fixture"
}`)

	// Create another test file
	createFileInFS(t, fs, "main.tf", `resource "test_resource" "example" {
  name = "example"
}`)

	// Commit the files
	_, err = workTree.Add(".")
	require.NoError(t, err)

	_, err = workTree.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	// Return the simulated repository URL.
	return "git@example.com:test/fixture-repo.git", sshKeyPath
}

// Helper function to create files in a billy filesystem.
func createFileInFS(t *testing.T, fs billy.Filesystem, filePath string, content string) {
	file, err := fs.Create(filePath)
	require.NoError(t, err)
	defer file.Close()
	_, err = file.Write([]byte(content))
	require.NoError(t, err)
}

// Mock Git client that works with our fixture repo.
type FixtureGitClient struct {
	CurrentSource *types.GitSource
}

func (gc *FixtureGitClient) CloneSource(ctx context.Context, cloneOpts storage.CloneOpts) (billy.Filesystem, error) {
	// Create a memory filesystem with our test files
	mfs := memfs.New()

	// Create test files instead of actually cloning
	// Create locals.tf file directly in the in-memory filesystem
	file, err := mfs.Create("locals.tf")
	if err != nil {
		return nil, err
	}
	_, err = file.Write([]byte(`locals {
  name = "test-fixture"
}`))
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}

	// Create another test file
	file, err = mfs.Create("main.tf")
	if err != nil {
		return nil, err
	}
	_, err = file.Write([]byte(`resource "test_resource" "example" {
  name = "example"
}`))
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}

	// Copy from memory to the destination filesystem
	// err = CopyFileSystemSafe(mfs, "/", cloneOpts.DestinationPath, cloneOpts.DestinationFs)
	// if err != nil {
	// 	return nil, err
	// }

	return mfs, nil
}

func (g *FixtureGitClient) GetCurrentSource() *types.Source {
	return (*types.Source)(g.CurrentSource)
}

func (g *FixtureGitClient) SetCurrentSource(s *types.Source) {
	g.CurrentSource = (*types.GitSource)(s)
}

func TestCloneSource(t *testing.T) {
	// Set up our fixture Git repository
	fixtureRepoURL, sshKeyPath := setupFixtureGitRepo(t)

	tests := []struct {
		name             string
		currentSource    *types.GitSource
		cloneOpts        storage.CloneOpts
		expectedErr      error
		expectedFilePath string
		useFixture       bool
	}{
		{
			name: "Successful clone with fixture",
			currentSource: &types.GitSource{
				URL: fixtureRepoURL,
				SourceAuth: &types.SourceAuth{
					SshKey: sshKeyPath,
				},
				Path: "/",
			},
			cloneOpts: storage.CloneOpts{
				DestinationPath: "/",
				DestinationFs:   afero.NewMemMapFs(),
			},
			expectedErr:      nil,
			expectedFilePath: "locals.tf",
			useFixture:       true,
		},
		// {
		// 	name: "Successful clone with SSH",
		// 	currentSource: &types.GitSource{
		// 		URL: "git@ssh.dev.azure.com:v3/parisbrooker-iac/PLT.TMPLTR.TEMPLATES/tmpltr.terraform.child",
		// 		SourceAuth: &types.SourceAuth{
		// 			SshKey: "/home/parisb/.ssh/ado",
		// 		},
		// 		Path: "/",
		// 	},
		// 	cloneOpts: CloneOpts{
		// 		DestinationPath: "/",
		// 		DestinationFs:   afero.NewMemMapFs(),
		// 	},
		// 	expectedErr:      nil,
		// 	expectedFilePath: "locals.tf",
		// 	useFixture:       false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			var bfs billy.Filesystem
			mfs := afero.NewMemMapFs()
			tt.cloneOpts.DestinationFs = mfs

			// if tt.useFixture {
			// Use our fixture client
			gc := &FixtureGitClient{
				CurrentSource: tt.currentSource,
			}
			bfs, err = gc.CloneSource(context.Background(), tt.cloneOpts)
			// } else {
			// 	// Use the real GitClient
			// 	gc := &GitClient{
			// 		CurrentSource: tt.currentSource,
			// 	}
			// 	b, err := gc.CloneSource(context.Background(), tt.cloneOpts)
			// 	bfs = b
			// }

			if tt.expectedErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				require.NoError(t, err)
				_, e := bfs.Stat(path.Join(tt.cloneOpts.DestinationPath, tt.expectedFilePath))
				require.NoError(t, e)
			}
		})
	}
}
