package parcello

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var _ Compressor = &ZipCompressor{}

// ErrSkipResource skips a particular file from processing
var ErrSkipResource = fmt.Errorf("Skip Resource Error")

// CompressorConfig controls how the code generation happens
type CompressorConfig struct {
	// Logger prints each step of compression
	Logger io.Writer
	// Filename is the name of the compressed bundle
	Filename string
	// IgnorePatterns provides a list of all files that has to be ignored
	IgnorePatterns []string
	// Recurive enables embedding the resources recursively
	Recurive bool
}

// ZipCompressor compresses content as GZip tarball
type ZipCompressor struct {
	// Config controls how the compression is made
	Config *CompressorConfig
}

// Compress compresses given source in tar.gz
func (e *ZipCompressor) Compress(ctx *CompressorContext) (*Bundle, error) {
	buffer := &bytes.Buffer{}
	count, err := e.write(buffer, ctx)

	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, nil
	}

	return &Bundle{
		Name:  e.Config.Filename,
		Body:  buffer.Bytes(),
		Count: count,
	}, nil
}

func (e *ZipCompressor) write(w io.Writer, ctx *CompressorContext) (int, error) {
	compressor := zip.NewWriter(w)
	if ctx.Offset > 0 {
		compressor.SetOffset(ctx.Offset)
	}

	count := 0

	err := ctx.FileSystem.Walk("/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		err = e.filter(path, info)

		switch err {
		case ErrSkipResource:
			return nil
		default:
			if err != nil {
				return err
			}
		}

		if err = e.walk(compressor, ctx.FileSystem, path, info); err != nil {
			return err
		}

		count = count + 1
		return nil
	})

	if err != nil {
		return count, err
	}

	_ = compressor.Flush()

	if ioErr := compressor.Close(); err == nil {
		err = ioErr
	}

	return count, err
}

func (e *ZipCompressor) walk(compressor *zip.Writer, fileSystem FileSystem, path string, info os.FileInfo) error {
	fmt.Fprintln(e.Config.Logger, fmt.Sprintf("Compressing '%s'", path))

	header, _ := zip.FileInfoHeader(info)
	header.Method = zip.Deflate
	header.Name = path

	writer, _ := compressor.CreateHeader(header)
	resource, err := fileSystem.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	defer func() {
		if ioErr := resource.Close(); err == nil {
			err = ioErr
		}
	}()

	_, err = io.Copy(writer, resource)
	return err
}

func (e *ZipCompressor) filter(path string, info os.FileInfo) error {
	if info == nil {
		return ErrSkipResource
	}

	if err := e.ignore(path, info); err != nil {
		return err
	}

	if !info.IsDir() {
		return nil
	}

	if !e.Config.Recurive && path != "." {
		return filepath.SkipDir
	}

	return ErrSkipResource
}

func (e *ZipCompressor) ignore(path string, info os.FileInfo) error {
	ignore := append(e.Config.IgnorePatterns, "*.go")

	for _, pattern := range ignore {
		matched, err := match(pattern, path, info.Name())

		if err != nil {
			return err
		}

		if !matched {
			continue
		}

		if info.IsDir() {
			return filepath.SkipDir
		}

		return ErrSkipResource
	}

	return nil
}
