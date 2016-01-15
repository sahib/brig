// Package fuse implements a FUSE layer for brig.
// Using it, a repository may be represented as a "normal" directory.
// There are three different structs in the FUSE API:
//
//     - fuse.Node  : A file or a directory (depending on it's type)
//	   - fuse.FS    : The filesystem. Used to find out the root node.
//     - fuse.Handle: An open file.
//
// This implementation offers File (a fuse.Node and fuse.Handle),
// Dir (fuse.Node) and FS (fuse.FS).
//
// Fuse will call the respective handlers if it needs information about your
// nodes. Each request handlers will usually get a `ctx` used to cancel
// operations, a request structure `req` with detailed query infos and
// a reponse structure `resp` where results are written. Usually the request
// handlers might return an error or a new node/handle/fs.
//
// Every request handle that may run for a long time should be
// made interruptable. Especially read and write operations should
// check the ctx.Done() channel passed to each request handler.
package fuse
