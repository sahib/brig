package endpoints

import (
	"html/template"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/csrf"
)

type IndexHandler struct {
}

func NewIndexHandler() *IndexHandler {
	return &IndexHandler{}
}

func (ih *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: Read from parcello.
	data, err := ioutil.ReadFile("gateway/templates/index.html")
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
	})

	if err != nil {
		jsonifyErrf(w, http.StatusInternalServerError, "could not execute template")
		return
	}
}
