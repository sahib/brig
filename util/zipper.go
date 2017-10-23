package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Tar packs all files in the directory pointed to by `root` and writes
// a gzipped and tarred version of it to `w`.
// The name of the archiv is set to `archiveName`.
func Tar(root, archiveName string, w io.Writer) error {
	gzw := gzip.NewWriter(w)
	gzw.Name = fmt.Sprintf(archiveName)
	gzw.Comment = ""
	gzw.ModTime = time.Now()

	tw := tar.NewWriter(gzw)
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		hdr := &tar.Header{
			Name: path[len(root):],
			Mode: 0600,
			Size: info.Size(),
		}

		if werr := tw.WriteHeader(hdr); err != nil {
			return werr
		}

		fd, err := os.Open(path)
		if err != nil {
			return err
		}

		defer Closer(fd)

		if _, err := io.Copy(tw, fd); err != nil {
			return err
		}

		return nil
	}

	if err := filepath.Walk(root, walker); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	return gzw.Close()
}

// Untar reads .tar data (from Tar()) from `r` and writes all files packed in it to `root`.
func Untar(r io.Reader, root string) error {
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		return fmt.Errorf("untar: %s exists or is not readable: %v", root, err)
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Create the necessary directory if necessary.
		fullPath := filepath.Join(root, hdr.Name)
		if oerr := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
			return oerr
		}

		// Overwrite the file in the target directory
		fd, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}

		if _, err := io.Copy(fd, tr); err != nil {
			return fd.Close()
		}

		if err := fd.Close(); err != nil {
			return err
		}
	}

	return gzr.Close()
}
