package util

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type archiveEntry struct {
	path string
	size int64
}

func addToTar(root string, entry archiveEntry, tw *tar.Writer) error {
	relPath := entry.path
	if len(entry.path) > len(root) {
		relPath = entry.path[len(root):]
		relPath = strings.TrimLeftFunc(relPath, func(r rune) bool {
			return r == filepath.Separator
		})
	}

	hdr := &tar.Header{
		Name: relPath,
		Mode: 0600,
		Size: entry.size,
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	fd, err := os.Open(entry.path) // #nosec
	if err != nil {
		return err
	}

	defer Closer(fd)

	_, err = io.Copy(tw, fd)
	return err
}

// Tar packs all files in the directory pointed to by `root` and writes
// a gzipped and tarred version of it to `w`.
// The name of the archiv is set to `archiveName`.
func Tar(root, archiveName string, w io.Writer) error {
	root = filepath.Clean(root)

	gzw := gzip.NewWriter(w)
	gzw.Name = fmt.Sprintf(archiveName)
	gzw.ModTime = time.Now()
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// First complete the walk to have a consistent set of files.
	// If we e.g. place the .tar file in the same directory, we
	// might iterate over itself, which will be unfortunate.
	entries := []archiveEntry{}
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		entries = append(entries, archiveEntry{
			path: path,
			size: info.Size(),
		})

		return nil
	}

	if err := filepath.Walk(root, walker); err != nil {
		return err
	}

	for _, entry := range entries {
		if err := addToTar(root, entry, tw); err != nil {
			return err
		}
	}

	return nil
}

// Untar reads .tar data (from Tar()) from `r` and writes all files packed in it to `root`.
func Untar(r io.Reader, root string) error {
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
