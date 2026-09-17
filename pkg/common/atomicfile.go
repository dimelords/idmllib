package common

import (
	"io"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes a file by streaming into a temporary file in the
// destination directory and renaming it into place once write has returned
// successfully. A failure at any point leaves the destination untouched, so a
// caller never ends up with a truncated or half-written document.
//
// The file is created with mode perm before the rename (subject to umask).
func WriteFileAtomic(path string, perm os.FileMode, write func(w io.Writer) error) (err error) {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()

	if err = write(tmp); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Chmod(perm); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
