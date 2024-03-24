package server

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterHealthCheck(router *mux.Router, path string) {
	router.HandleFunc(path, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
	}).Methods(http.MethodGet)
}
