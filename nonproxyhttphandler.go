package main

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
)

var (
	HTMLTemplate *template.Template
)

func init() {
	HTMLTemplate = template.New("")
	files, _ := AssetDir("tmpl/html")
	for _, f := range files {
		template.Must(
			HTMLTemplate.New(
				strings.TrimSuffix(f, ".tmpl"),
			).Parse(
				string(MustAsset("tmpl/html/" + f)),
			),
		)
	}
}

func NewNonProxyHTTPHandler(pac *Pac) http.Handler {
	router := httprouter.New()
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		d := struct {
			Name         string
			Version      string
			PacFilename  string
			KnownProxies map[string]*PacConn
		}{
			Name:         Name,
			Version:      Version,
			PacFilename:  pac.PacFilename(),
			KnownProxies: pac.ConnService.KnownProxies(),
		}
		HTMLTemplate.ExecuteTemplate(w, "home", d)
	})
	router.GET("/proxy.pac", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		t := r.URL.Query().Get("t")
		if t == "1" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig; charset=utf-8")
		}
		w.Write(pac.PacConfiguration())
	})
	router.GET("/lookup-proxy", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		urlParam := r.URL.Query().Get("url")
		if urlParam == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(urlParam, "http://") && !strings.HasPrefix(urlParam, "https://") {
			urlParam = "http://" + urlParam
		}
		urlURL, err := url.Parse(urlParam)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		pacResult, err := pac.CallFindProxyForURLFromURL(urlURL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError)+"\n\n"+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(pacResult))
	})
	return router
}
