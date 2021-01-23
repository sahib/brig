// The purpose of this module is to implement a hint system.
// brig uses this to let users configure what files and directories
// should be encrypted and/or compressed. Some folders might also
// get none of the both.
//
// We store the hints in the repository as yaml file and store
// it in in-memory in a trie during runtime.
package hints
