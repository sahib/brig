// Package nodes implements all nodes and defines basic operations on it.
//
// It however does not implement any specific database scheme, nor
// are any operations implemented that require knowledge of other nodes.
// If knowledge about other nodes is required, the Linker interface needs
// to be fulfilled by a higher level.
//
// The actual core of brig is built upon this package.
// Any changes here should thus be well thought through.
package nodes
