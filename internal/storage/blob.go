package storage

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
)

// NewBlobClient creates a new BlobClient
func NewBlobClient() *BlobClient {
	return &BlobClient{}
}

type BlobClient struct {
	CurrentSource *types.BlobSource
}

func (bc *BlobClient) Clone(ctx context.Context) (billy.Filesystem, error) {
	// TODO: Implement git storage source cloning/downloading
	return nil, nil
}

func (bc *BlobClient) CloneSource(ctx context.Context, cloneOpts CloneOpts) (billy.Filesystem, error) {
	// TODO: Implement blob storage source cloning/downloading
	return nil, nil
}

func (bc *BlobClient) SetSource(s *types.Source) {
	bc.CurrentSource = (*types.BlobSource)(s)
}
