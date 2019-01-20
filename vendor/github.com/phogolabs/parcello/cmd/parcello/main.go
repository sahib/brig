// Command Line Interface of Embedo.
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/phogolabs/parcello"
	"github.com/urfave/cli"
)

const (
	// ErrCodeArg is returned when an invalid argument is passed to CLI
	ErrCodeArg = 101
)

func main() {
	app := &cli.App{
		Name:                 "parcello",
		HelpName:             "parcello",
		Usage:                "Golang Resource Bundler and Embedder",
		UsageText:            "parcello [global options]",
		Version:              "0.8",
		BashComplete:         cli.DefaultAppComplete,
		EnableBashCompletion: true,
		Writer:               os.Stdout,
		ErrWriter:            os.Stderr,
		Action:               run,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "quiet, q",
				Usage: "disable logging",
			},
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "embed or bundle the resources recursively",
			},
			cli.StringFlag{
				Name:  "resource-dir, d",
				Usage: "path to directory",
				Value: ".",
			},
			cli.StringFlag{
				Name:  "bundle-path, b",
				Usage: "path to the bundle directory or binary",
				Value: ".",
			},
			cli.StringFlag{
				Name:  "resource-type, t",
				Usage: "resource type. (supported: bundle, source-code)",
				Value: "source-code",
			},
			cli.StringSliceFlag{
				Name:  "ignore, i",
				Usage: "ignore file name",
			},
			cli.BoolTFlag{
				Name:  "include-docs",
				Usage: "include API documentation in generated source code",
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

func run(ctx *cli.Context) error {
	rType := ctx.String("resource-type")

	switch strings.ToLower(rType) {
	case "source-code":
		return embed(ctx)
	case "bundle":
		return bundle(ctx)
	default:
		err := fmt.Errorf("Invalid resource type '%s'", rType)
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}
}

func embed(ctx *cli.Context) error {
	resourceDir, err := filepath.Abs(ctx.String("resource-dir"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	bundlePath, err := filepath.Abs(ctx.String("bundle-path"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	_, packageName := filepath.Split(bundlePath)

	embedder := &parcello.Embedder{
		Logger:     logger(ctx),
		FileSystem: parcello.Dir(resourceDir),
		Composer: &parcello.Generator{
			FileSystem: parcello.Dir(bundlePath),
			Config: &parcello.GeneratorConfig{
				Package:     packageName,
				InlcudeDocs: ctx.BoolT("include-docs"),
			},
		},
		Compressor: &parcello.ZipCompressor{
			Config: &parcello.CompressorConfig{
				Logger:         logger(ctx),
				Filename:       "resource",
				IgnorePatterns: ctx.StringSlice("ignore"),
				Recurive:       ctx.Bool("recursive"),
			},
		},
	}

	if err := embedder.Embed(); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	return nil
}

func bundle(ctx *cli.Context) error {
	resourceDir, err := filepath.Abs(ctx.String("resource-dir"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	bundlePath, err := filepath.Abs(ctx.String("bundle-path"))
	if err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	bundler := &parcello.Bundler{
		Logger:     logger(ctx),
		FileSystem: parcello.Dir(resourceDir),
		Compressor: &parcello.ZipCompressor{
			Config: &parcello.CompressorConfig{
				Logger:         logger(ctx),
				Filename:       "resource",
				IgnorePatterns: ctx.StringSlice("ignore"),
				Recurive:       ctx.Bool("recursive"),
			},
		},
	}

	bundleDir, bundleName := filepath.Split(bundlePath)

	bctx := &parcello.BundlerContext{
		Name:       bundleName,
		FileSystem: parcello.Dir(bundleDir),
	}

	if err := bundler.Bundle(bctx); err != nil {
		return cli.NewExitError(err.Error(), ErrCodeArg)
	}

	return nil
}

func logger(ctx *cli.Context) io.Writer {
	if ctx.GlobalBool("quiet") {
		return ioutil.Discard
	}

	return os.Stdout
}
