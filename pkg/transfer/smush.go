package transfer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

/*
File for compressing directories into .zip, gzip, or .tar files before or after transfer
*/

// checks whether a file path is vulnerable to zip slip.
//
// see: https://security.snyk.io/research/zip-slip-vulnerability
func ValidPath(filePath string, dest string) bool {
	return strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator))
}

func Zip(sourceDir string, destArchive string) error {
	file, err := os.Create(destArchive)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	w := zip.NewWriter(file)
	defer w.Close()

	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			dirpath := fmt.Sprintf("%s%c", path, os.PathSeparator)
			_, err = w.Create(dirpath)
			if err != nil {
				return err
			}
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		f, err := w.Create(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}
		return nil
	}

	// walk directory and zip all files and sub directories
	err = filepath.Walk(sourceDir, walker)
	if err != nil {
		panic(err)
	}
	return nil
}

// from: https://stackoverflow.com/questions/20357223/easy-way-to-unzip-file
func Unzip(src string, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// check for ZipSlip (Directory traversal)
		if !ValidPath(path, dest) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}
