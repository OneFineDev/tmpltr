package storage

import (
	"context"

	"github.com/OneFineDev/tmpltr/internal/types"
	"github.com/go-git/go-billy/v5"
)

type BlobClient struct {
	CurrentSource *types.BlobSource
}

// NewBlobClient creates a new BlobClient.
func NewBlobClient() *BlobClient {
	return &BlobClient{}
}

func (bc *BlobClient) Clone(_ context.Context) (billy.Filesystem, error) {
	// TODO: Implement git storage source cloning/downloading
	return nil, ErrNotImplemented
}

func (bc *BlobClient) SetSource(s *types.Source) {
	bc.CurrentSource = (*types.BlobSource)(s)
}
