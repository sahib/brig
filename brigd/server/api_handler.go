package server

import "github.com/disorganizer/brig/brigd/capnp"

type apiHandler struct {
	metaHandler
	fsHandler
	vcsHandler
}

func newApiHandler(base *base) *apiHandler {
	ah := &apiHandler{}

	// Divide the implementation of all handlers over the respective areas:
	ah.metaHandler.base = base
	ah.fsHandler.base = base
	ah.vcsHandler.base = base
	return ah
}

func (ah *apiHandler) Version(call capnp.API_version) error {
	call.Results.SetVersion(1)
	return nil
}
