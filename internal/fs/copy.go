package fs

import (
	"io"
	"os"

	"github.com/pkg/errors"
)

// CopyFile copies the source path to the destination path.
func CopyFile(src, dest string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return errors.Wrapf(err, "cannot stat %s", src)
	}
	srcSize := srcStat.Size()

	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "cannot open %s", src)
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return errors.Wrapf(err, "cannot create %s", dest)
	}
	defer destFile.Close()

	written, err := io.Copy(destFile, srcFile)
	if err != nil {
		return errors.Wrapf(err, "unable to copy %s to %s", src, dest)
	}
	if written != srcSize {
		return errors.Wrapf(err, "copied the wrong number of bytes (%s instead of %s) from %s to %s",
			written, srcSize, src, dest)
	}

	err = destFile.Sync()
	if err != nil {
		return errors.Wrapf(err, "unable to flush %s to disk", dest)
	}

	return nil
}
