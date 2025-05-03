package storage

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
)

type FileClient struct {
	CurrentSource *types.FileSource
}

// NewFileClient creates a new FileClient.
func NewFileClient() *FileClient {
	return &FileClient{}
}

func (fc *FileClient) Clone(_ context.Context) (billy.Filesystem, error) {
	// TODO: Implement git storage source cloning/downloading
	return nil, ErrNotImplemented
}

// SetSource sets the current source.
func (fc *FileClient) SetSource(s *types.Source) {
	fc.CurrentSource = (*types.FileSource)(s)
}
