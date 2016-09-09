package store

import "testing"

// TODO: move to test
func (d *Directory) Equal(o *Directory) bool {
	return true &&
		d.name == o.name &&
		d.hash.Equal(o.hash) &&
		d.size == o.size &&
		d.modTime.Equal(o.modTime)
}

func TestDirectoryCreating(t *testing.T) {
	// TODO: write.
}
