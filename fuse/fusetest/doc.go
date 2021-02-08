// Package fusetest offers an easy way to test our fuse code.
//
// What this does is start another process with an HTTP server in it.
// Beside the HTTP server the fuse mount is mounted at a specified path,
// with the specified options. A client can connect to the server and control
// it and/or a client program can access files in the fuse mount.
//
// The reason why this is another process is an issue with Go:
// When serving and accessing the FUSE mount in the same we might enter
// an unrecoverable deadlock where file I/O related syscalls get stuck
// because the go routine that serve this syscall live in the same process
// but do not get called because only N parallel go routines can be run.
//
// Reference:
//
// * https://github.com/bazil/fuse/issues/264#issuecomment-727269770
// * https://github.com/sahib/brig/pull/77#issuecomment-754831080
//
// bazil/fuse offers an spawntest utility that does something very similar,
// but we also this package in benchmarks and spawntest expects to be called
// from tests and thus requires the testing package.
package fusetest
