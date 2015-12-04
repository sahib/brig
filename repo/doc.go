// Package repo offers function for creating and loading a brig repository.
//
// The repository looks like this:
//
// /path/to/repo
// └── .brig
//     ├── config
//     ├── index.bolt[.minilock]
//     ├── master.key[.minilock]
//     └── ipfs
//         └── ...
//
// Directly after `init`, the index and key files will be still encrypted
// with minilock. `open` will use the user's password to decrypt those.
// `close` reverses this by encrypting them again.
//
// The `Repository` structure aids in accessing all those files and offers
// individual apis for them (like `Store` for reading/writing the index).
package repo
