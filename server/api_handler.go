package server

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
