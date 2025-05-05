//go:build !integration

package storage_test

import (
	"errors"
	"testing"

	"github.com/OneFineDev/tmpltr/internal/storage"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeFs_CopyFileSystemSafe(t *testing.T) {
	// Arrange
	tests := []struct {
		name          string
		setupSourceFs func() billy.Filesystem
		setupDestFs   func() afero.Fs
		root          string
		dest          string
		expectedError error
		verify        func(t *testing.T, destFs afero.Fs)
	}{
		{
			name: "successfully copies files and directories",
			setupSourceFs: func() billy.Filesystem {
				fs := memfs.New()
				fs.MkdirAll("/src/dir", 0755)
				file, _ := fs.Create("/src/dir/file.txt")
				file.Write([]byte("content"))
				return fs
			},
			setupDestFs: func() afero.Fs { //nolint:gocritic // For sig consistency
				return afero.NewMemMapFs()
			},
			root:          "/src",
			dest:          "/dest",
			expectedError: nil,
			verify: func(t *testing.T, destFs afero.Fs) {
				_, err := destFs.Stat("/dest/src/dir")
				require.NoError(t, err)
				content, err := afero.ReadFile(destFs, "/dest/src/dir/file.txt")
				require.NoError(t, err)
				assert.Equal(t, "content", string(content))
			},
		},
		{
			name: "returns error when source directory does not exist",
			setupSourceFs: func() billy.Filesystem { //nolint:gocritic // For sig consistency
				return memfs.New()
			},
			setupDestFs: func() afero.Fs { //nolint:gocritic // For sig consistency
				return afero.NewMemMapFs()
			},
			root:          "/nonexistent",
			dest:          "/dest",
			expectedError: errors.New("file does not exist"),
			verify:        func(*testing.T, afero.Fs) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			sourceFs := tt.setupSourceFs()
			destFs := tt.setupDestFs()
			safeFs := &storage.SafeFs{Fs: destFs}

			// Act
			err := safeFs.CopyFileSystemSafe(sourceFs, tt.root, tt.dest)

			// Assert
			if tt.expectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
			tt.verify(t, destFs)
		})
	}
}
