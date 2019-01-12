package endpoints

import (
	"html/template"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"
	"github.com/phogolabs/parcello"

	// Include static resources:
	_ "github.com/sahib/brig/gateway/templates"
)

type IndexHandler struct {
	State
}

func NewIndexHandler(s State) *IndexHandler {
	return &IndexHandler{State: s}
}

func (ih *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mgr := parcello.ManagerAt("/")
	fd, err := mgr.Open("index.html")
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "no index.html")
		return
	}

	defer fd.Close()

	data, err := ioutil.ReadAll(fd)
	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "could not load template: %v", err)
		return
	}

	t, err := template.New("index").Parse(string(data))
	if err != nil {
		log.Errorf("could not parse template: %v", err)
		jsonifyErrf(w, http.StatusInternalServerError, "template contains errors")
		return
	}

	err = t.Execute(w, map[string]interface{}{
		"csrfToken": csrf.Token(r),
		"wsAddr":    "wss://" + r.Host + "/events",
	})

	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "could not execute template")
		return
	}
}
