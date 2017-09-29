package server

import "github.com/disorganizer/brig/brigd/capnp"

type fsHandler struct {
	base
}

func (fh *fsHandler) Stage(call capnp.FS_stage) error {
	return nil
}
