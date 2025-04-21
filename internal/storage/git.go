package storage

import (
	"context"
	"strings"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

func init() {
	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}
}

type GitClient struct {
	CurrentSource *types.GitSource
}

func NewGitClient() *GitClient {
	return &GitClient{}
}

func (gc *GitClient) CloneSource(ctx context.Context, cloneOpts CloneOpts) (billy.Filesystem, error) {
	// File systems #TODO: Implement local save
	mfs := memfs.New()

	// Check that auth method matches transport
	isSshTransport := strings.Contains(gc.CurrentSource.URL, "ssh")

	if isSshTransport && gc.CurrentSource.SshKey == "" {
		return nil, &TransportAuthMismatchError{
			ExpectedAuthMethod: "ssh",
			Url:                gc.CurrentSource.URL,
		}
	}

	if !isSshTransport && gc.CurrentSource.Pat == "" {
		return nil, &TransportAuthMismatchError{
			ExpectedAuthMethod: "PAT",
			Url:                gc.CurrentSource.URL,
		}
	}

	// build the expected git auth
	var gitAuth transport.AuthMethod
	switch isSshTransport {
	case true:
		publicKeys, err := ssh.NewPublicKeysFromFile("git", gc.CurrentSource.SshKey, "")
		if err != nil {
			return nil, &SshKeyError{
				SshKeyPath: gc.CurrentSource.SshKey,
				OpErr:      err,
			}
		}
		gitAuth = publicKeys
	default:
		gitAuth = &http.BasicAuth{
			Username: gc.CurrentSource.UserName,
			Password: gc.CurrentSource.Pat,
		}
	}

	gitOpts := &git.CloneOptions{
		URL:          gc.CurrentSource.URL,
		Auth:         gitAuth,
		Depth:        1,
		SingleBranch: true,
	}

	stg := memory.NewStorage()
	_, err := git.Clone(stg, mfs, gitOpts)
	if err != nil {
		return nil, err
	}

	rerooted, err := mfs.Chroot(gc.CurrentSource.Path)
	if err != nil {
		return nil, err
	}

	return rerooted, nil

	// err = CopyFileSystem(rerooted, "/", cloneOpts.DestinationPath, cloneOpts.DestinationFs)
	// if err != nil {
	// 	return err
	// }

	// return nil
}

func (g *GitClient) GetCurrentSource() *types.Source {
	return (*types.Source)(g.CurrentSource)
}

func (g *GitClient) SetCurrentSource(s *types.Source) {
	g.CurrentSource = (*types.GitSource)(s)
}
