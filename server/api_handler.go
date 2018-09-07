package server

type apiHandler struct {
	repoHandler
	fsHandler
	vcsHandler
	netHandler
}

func newApiHandler(base *base) *apiHandler {
	return &apiHandler{
		repoHandler: repoHandler{base},
		netHandler:  netHandler{base},
		vcsHandler:  vcsHandler{base},
		fsHandler:   fsHandler{base},
	}
}
