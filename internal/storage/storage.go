package storage

import (
	"errors"

	"github.com/spf13/afero"
)

type CloneOpts struct {
	DestinationPath string
	DestinationFs   afero.Fs
}

var ErrNotImplemented error = errors.New("not implemented")
