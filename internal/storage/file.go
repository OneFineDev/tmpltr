package storage

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
)

type FileClient struct {
	CurrentSource *types.Source
}

func (fc *FileClient) CloneSource(ctx context.Context, cloneOpts CloneOpts) (billy.Filesystem, error) {
	// TODO: Implement git storage source cloning/downloading
	return nil, nil
}

// GetCurrentSource returns the currently set source
func (fc *FileClient) GetCurrentSource() *types.Source {
	// TODO: Return the proper *types.Source when fully implemented
	return fc.CurrentSource
}

// SetCurrentSource sets the current source
func (fc *FileClient) SetCurrentSource(source *types.Source) {
	// TODO: Validate that source is a *types.Source when fully implemented
	fc.CurrentSource = source
}

// NewFileClient creates a new FileClient
func NewFileClient() *FileClient {
	return &FileClient{}
}
