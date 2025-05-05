package storage

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/util"
	"github.com/spf13/afero"
)

type SafeFs struct {
	mu sync.Mutex
	Fs afero.Fs
}

// CopyFileSystemSafe recursively walks a directory and copies its contents.
func (sf *SafeFs) CopyFileSystemSafe(fs billy.Filesystem, root string, dest string) error {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return util.Walk(fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		destPath := filepath.Join(dest, path)

		if info.IsDir() {
			e := sf.Fs.MkdirAll(destPath, 0775) //nolint:mnd
			if e != nil {
				return e
			}
		} else {
			e := copyFile(fs, path, destPath, sf.Fs)
			if e != nil {
				return e
			}
		}
		return nil
	})
}

// copyFile copies a file.
func copyFile(fs billy.Filesystem, src, dest string, localFs afero.Fs) error {
	sourceFile, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := localFs.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}
