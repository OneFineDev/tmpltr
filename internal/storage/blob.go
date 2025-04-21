package storage

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
)

type BlobClient struct {
	CurrentSource *types.Source
}

func (bc *BlobClient) CloneSource(ctx context.Context, cloneOpts CloneOpts) (billy.Filesystem, error) {
	// TODO: Implement blob storage source cloning/downloading
	return nil, nil
}

// GetCurrentSource returns the currently set source
func (bc *BlobClient) GetCurrentSource() *types.Source {
	// TODO: Return the proper *types.Source when fully implemented
	return bc.CurrentSource
}

// SetCurrentSource sets the current source
func (bc *BlobClient) SetCurrentSource(source *types.Source) {
	// TODO: Validate that source is a *types.Source when fully implemented
	bc.CurrentSource = source
}

// NewBlobClient creates a new BlobClient
func NewBlobClient() *BlobClient {
	return &BlobClient{}
}
