# Parcello

[![Documentation][godoc-img]][godoc-url]
![License][license-img]
[![Build Status][travis-img]][travis-url]
[![Coverage][codecov-img]][codecov-url]
[![Go Report Card][report-img]][report-url]

*Golang Resource Bundler*

[![Parcel][parcello-img]][parcello-url]

## Overview

Parcello is a simple resource manager for Golang that allows embedding assets
like SQL, bash scripts and images. That allows easy release management by
deploying just a single binary rather than many files.

## Roadmap

Note that we may introduce breaking changes until we reach v1.0.

- [x] Rename the tool in order not to clash with [parcel-bundler](https://github.com/parcel-bundler/parcel)
- [x] Support [http.FileSystem](https://golang.org/pkg/net/http/#FileSystem)
- [x] Bundle resource as ZIP archive in the end of built Golang binary
- [ ] ~~Support embedded COFF resources~~ (postponed until we accomplish a spike that works on all platforms)

## Installation

#### GitHub

```console
$ go get -u github.com/phogolabs/parcello
$ go install github.com/phogolabs/parcello/cmd/parcello
```
#### Homebrew (for Mac OS X)

```console
$ brew tap phogolabs/tap
$ brew install parcello
```

## Usage

You can use the parcello command line interface to bundle the desired resources
recursively:

```console
$ parcello -r -d <resource_dir_source> -b <bundle_dir_destination>
```

However, the best way to use the tool is via `go generate`. In order to embed all
resource in particular directory, you should make it a package that has the
following comment:

```golang
// Package database contains the database artefacts of GOM as embedded resource
package database

//go:generate parcello -r
```

When you run:

```console
$ go generate ./...
```

The tools will create a `resource.go` file that contains
all embedded resource in that directory and its
subdirectories as `zip` archive which is registered in
[parcello.ResourceManager](https://github.com/phogolabs/parcello/blob/master/common.go#L6).

You can read the content in the following way:

```golang
// Import the package that includes 'resource.go'
import _ "database"

file, err := parcello.Open("your_sub_directory_name/your_file_name")
```

The `parcello` package provides an abstraction of
[FileSystem](https://godoc.org/github.com/phogolabs/parcello#FileSystem)
interface:

```golang
// FileSystem provides primitives to work with the underlying file system
type FileSystem interface {
	// A FileSystem implements access to a collection of named files.
	http.FileSystem
	// Walk walks the file tree rooted at root, calling walkFn for each file or
	// directory in the tree, including root.
	Walk(dir string, fn filepath.WalkFunc) error
	// OpenFile is the generalized open call; most users will use Open
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}
```

That is implemented by the following:

- [parcello.ResourceManager](https://godoc.org/github.com/phogolabs/parcello#ResourceManager) which provides an access to the bundled resources.
- [parcello.Dir](https://godoc.org/github.com/phogolabs/parcello#Dir) which provides an access to the underlying file system.

That allows easy replacement of the file system with the bundled resources and
vice versa.

If you want to work in dev mode, you should set the following environment
variables before you start your application:

```console
$ export PARCELLO_DEV_ENABLED=1
$ # if the application resource directory is different than the current working directory
$ export PARCELLO_RESOURCE_DIR=./public
```

Note that downsides of this resource embedding approach are that your compile
time may increase significantly.

If you have such a issue, you can bundle the resource at the end of your binary
as zip archive. You can do this via `parcello` CLI:

```console
$ go build your_binary
$ parcello -r -d <resource_dir_source> -b <path_to_your_binary> -t bundle
```

## Command Line Interface

```console
$ parcello -h

NAME:
   parcello - Golang Resource Bundler and Embedder

USAGE:
   parcello [global options]

VERSION:
   0.8

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --bundle-path value, -b value    path to the bundle directory or binary (default: ".")
   --ignore value, -i value         ignore file name
   --include-docs                   include API documentation in generated source code
   --quiet, -q                      disable logging
   --recursive, -r                  embed or bundle the resources recursively
   --resource-dir value, -d value   path to directory (default: ".")
   --resource-type value, -t value  resource type. (supported: bundle, source-code) (default: "source-code")
   --help, -h                       show help
   --version, -v                    print the version
```

## Example

You can check working [example](example).

## Contributing

We are welcome to any contributions. Just fork the
[project](https://github.com/phogolabs/parcello).

*logo made by [Good Wave][logo-author-url] [CC 3.0][logo-license]*

[report-img]: https://goreportcard.com/badge/github.com/phogolabs/parcello
[report-url]: https://goreportcard.com/report/github.com/phogolabs/parcello
[logo-author-url]: https://www.flaticon.com/authors/good-ware
[logo-license]: http://creativecommons.org/licenses/by/3.0/
[parcello-url]: https://github.com/phogolabs/parcello
[parcello-img]: doc/img/logo.png
[codecov-url]: https://codecov.io/gh/phogolabs/parcello
[codecov-img]: https://codecov.io/gh/phogolabs/parcello/branch/master/graph/badge.svg
[travis-img]: https://travis-ci.org/phogolabs/parcello.svg?branch=master
[travis-url]: https://travis-ci.org/phogolabs/parcello
[parcello-url]: https://github.com/phogolabs/parcello
[godoc-url]: https://godoc.org/github.com/phogolabs/parcello
[godoc-img]: https://godoc.org/github.com/phogolabs/parcello?status.svg
[license-img]: https://img.shields.io/badge/license-MIT-blue.svg
[software-license-url]: LICENSE
