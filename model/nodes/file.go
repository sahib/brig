package nodes

import "time"

// File represents a single file in the repository.
// It stores all metadata about it and links to the actual data.
type File struct {
	Base

	key     []byte
	parent  string
	size    uint64
	modTime time.Time
	id      uint64
}

// func NewEmptyFile(fs *FS, parent *Directory, name string) (*File, error) {
// 	id, err := fs.NextID()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	file := &File{
// 		name:    name,
// 		id:      id,
// 		modTime: time.Now(),
// 		fs:      fs,
// 		parent:  parent.Path(),
// 	}
//
// 	return file, nil
// }
