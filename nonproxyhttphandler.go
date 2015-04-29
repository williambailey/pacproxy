package main

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewNonProxyHTTPHandler(pac *Pac) http.Handler {
	router := httprouter.New()
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s v%s %s\n", Name, Version, pac.PacFilename())
	})
	router.GET("/wpad.dat", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.Write(pac.PacConfiguration())
	})
	return router
}
