package storage

import (
	"github.com/spf13/afero"
)

type StorageAuth interface {
	AuthType() string
}

type CloneOpts struct {
	DestinationPath string
	DestinationFs   afero.Fs
}
