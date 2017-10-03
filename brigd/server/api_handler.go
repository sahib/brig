package server

import "github.com/disorganizer/brig/brigd/capnp"

type apiHandler struct {
	metaHandler
	fsHandler
}

func newApiHandler(base *base) *apiHandler {
	ah := &apiHandler{}
	ah.metaHandler.base = base
	ah.fsHandler.base = base
	return ah
}

func (ah *apiHandler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
