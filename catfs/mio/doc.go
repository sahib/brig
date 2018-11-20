// Package mio (short for memory input/output) implements the layered io stack
// of brig. This includes currently three major parts:
//
// - encrypt  - Encryption and Decryption layer with seeking support.
// - compress - Seekable Compression and Decompression with exchangable algorithms.
// - overlay  - In-Memory write overlay over a io.Reader with seek support.
//
// This package itself contains utils that stack those on top of each of other
// in an already usable fashion.
package mio
