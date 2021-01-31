// Package hints implements a hint system for streaming.
// Brig uses this to let users configure what files and directories
// should be encrypted and/or compressed. Some folders might also
// get none of the both. In the latter case we call the stream "raw".
//
// We store the hints in the repository as yaml file and store
// it in in-memory in a trie during runtime.
package hints
