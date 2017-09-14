package catfs

import (
	"io"
	"sync"

	"github.com/disorganizer/brig/catfs/db"
)

type FS struct {
	mu sync.Mutex

	kv  db.Database
	lkr *Linker
	bk  FsBackend
}

func NewFilesystem(dbPath, owner string) (*FS, error) {
	return &FS{}, nil
}

func (fs *FS) Close() error {
	return nil
}

func (fs *FS) Export(w io.Writer) error {
	return nil
}

func (fs *FS) Import(r io.Reader) error {
	return nil
}

/////////////////////
// CORE OPERATIONS //
/////////////////////

func (fs *FS) Move(src, dst string) error {
	return nil
}

func (fs *FS) Mkdir(path string, createParents bool) error {
	return nil
}

func (fs *FS) Remove(path string) error {
	return nil
}

type NodeInfo struct {
	Path  string
	Type  int
	Size  uint64
	Inode uint64
}

func (fs *FS) Stat(path string) (*NodeInfo, error) {
	return nil, nil
}

/////////////////////
// SYNC OPERATIONS //
/////////////////////

func (fs *FS) Sync(remote *FS) error {
	return nil
}

type Diff struct {
	Ignored  map[string]*NodeInfo
	Removed  map[string]*NodeInfo
	Added    map[string]*NodeInfo
	Merged   map[string]*NodeInfo
	Conflict map[string]*NodeInfo
}

func (fs *FS) Diff(remot *FS) (*Diff, error) {
	return nil, nil
}

////////////////////////
// PINNING OPERATIONS //
////////////////////////

func (fs *FS) Pin(path string) error {
	return nil
}

func (fs *FS) Unpin(path string) error {
	return nil
}

func (fs *FS) IsPinned(path string) (bool, error) {
	return false, nil
}

////////////////////////
// STAGING OPERATIONS //
////////////////////////

func (fs *FS) Touch(path string) error {
	return nil
}

func (fs *FS) StageFromReader(path string, r io.Reader) error {
	return nil
	// path = prefixSlash(path)

	// fs.mu.Lock()
	// defer fs.mu.Unlock()

	// // Control how many bytes are written to the encryption layer:
	// sizeAcc := &util.SizeAccumulator{}
	// teeR := io.TeeReader(r, sizeAcc)

	// key := make([]byte, 32)
	// if _, err := io.ReadFull(rand.Reader, key); err != nil {
	// 	return err
	// }

	// stream, err := NewFileReader(key, teeR, compressAlgo)
	// if err != nil {
	// 	return err
	// }

	// hash, err := st.backend.Add(stream)
	// if err != nil {
	// 	return err
	// }

	// if err := st.backend.Pin(hash); err != nil {
	// 	return err
	// }

	// owner, err := lkr.Owner()
	// if err != nil {
	// 	return err
	// }

	// // if _, err := stageFile(st.fs, repoPath, hash, sizeAcc.Size(), owner.ID(), key); err != nil {
	// // 	return err
	// // }

	// return nil
}
