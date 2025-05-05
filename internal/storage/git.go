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

func init() { //nolint:gochecknoinits // needed
	transport.UnsupportedCapabilities = []capability.Capability{ //nolint:reassign // needed
		capability.ThinPack,
	}
}

type GitClient struct {
	CurrentSource *types.GitSource
}

func NewGitClient() *GitClient {
	return &GitClient{}
}

func (gc *GitClient) Clone(ctx context.Context) (billy.Filesystem, error) {
	mfs := memfs.New()

	// Check that auth method matches transport
	isSSHTransport := strings.Contains(gc.CurrentSource.URL, "ssh")

	if isSSHTransport && gc.CurrentSource.SSHKey == "" {
		return nil, &TransportAuthMismatchError{
			ExpectedAuthMethod: "ssh",
			URL:                gc.CurrentSource.URL,
		}
	}

	if !isSSHTransport && gc.CurrentSource.Pat == "" {
		return nil, &TransportAuthMismatchError{
			ExpectedAuthMethod: "PAT",
			URL:                gc.CurrentSource.URL,
		}
	}

	// build the expected git auth
	var gitAuth transport.AuthMethod
	switch isSSHTransport {
	case true:
		publicKeys, err := ssh.NewPublicKeysFromFile("git", gc.CurrentSource.SSHKey, "")
		if err != nil {
			return nil, &SSHKeyError{
				SSHKeyPath: gc.CurrentSource.SSHKey,
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
	_, err := git.CloneContext(ctx, stg, mfs, gitOpts)
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

func (gc *GitClient) SetSource(s *types.Source) {
	gc.CurrentSource = (*types.GitSource)(s)
}
