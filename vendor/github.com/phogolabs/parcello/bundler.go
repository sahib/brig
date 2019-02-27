package parcello

import (
	"fmt"
	"io"
	"os"
)

// BundlerContext the context of this bundler
type BundlerContext struct {
	// Name of the binary
	Name string
	// FileSystem represents the underlying file system
	FileSystem FileSystem
}

// Bundler bundles the resources to the provided binary
type Bundler struct {
	// Logger prints each step of compression
	Logger io.Writer
	// Compressor compresses the resources
	Compressor Compressor
	// FileSystem represents the underlying file system
	FileSystem FileSystem
}

// Bundle bundles the resources to the provided binary
func (e *Bundler) Bundle(ctx *BundlerContext) error {
	file, err := ctx.FileSystem.OpenFile(ctx.Name, os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}

	defer file.Close()

	finfo, ferr := file.Stat()
	if ferr != nil {
		return ferr
	}

	if finfo.IsDir() {
		return fmt.Errorf("'%s' is not a regular file", ctx.Name)
	}

	cctx := &CompressorContext{
		FileSystem: e.FileSystem,
		Offset:     finfo.Size(),
	}

	fmt.Fprintf(e.Logger, "Bundling resource(s) at '%s'\n", ctx.Name)
	bundle, cerr := e.Compressor.Compress(cctx)
	if cerr != nil {
		return cerr
	}

	if _, err = file.Write(bundle.Body); err != nil {
		return err
	}

	fmt.Fprintf(e.Logger, "Bundled %d resource(s) at '%s'\n", bundle.Count, ctx.Name)
	return nil
}
