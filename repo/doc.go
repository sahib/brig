// Package repo offers function for creating and loading a brig repository.
//
// The repository looks like this:
//
// /path/to/repo
// └── .brig
//     ├── config
//     ├── remotes.yml[.locked]
//     ├── index.bolt[.locked]
//     ├── master.key[.locked]
//     └── ipfs
//         └── ...
//
// Directly after `init`, the index and key files will be still encrypted
// with AES-GCM. `open` will use the user's password to decrypt those.
// `close` reverses this by encrypting them again.
//
// The `Repository` structure aids in accessing all those files and offers
// individual apis for them (like `Store` for reading/writing the index).
package repo
