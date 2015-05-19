package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	"github.com/elazarl/go-bindata-assetfs"
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
	router.Handler(
		"GET",
		"/pacproxy/*filepath",
		http.StripPrefix(
			"/pacproxy",
			http.FileServer(
				&assetfs.AssetFS{
					Asset:    Asset,
					AssetDir: AssetDir,
					Prefix:   "htdocs",
				},
			),
		),
	)
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
		d := struct {
			Name        string
			Version     string
			PacFilename string
		}{
			Name:        Name,
			Version:     Version,
			PacFilename: pac.PacFilename(),
		}
		HTMLTemplate.ExecuteTemplate(w, "home", d)
	})
	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/html")
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

		HTMLTemplate.ExecuteTemplate(w, "status", d)
	})
	router.GET("/wpad.dat", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.Write(
			bytes.Replace(
				MustAsset("wpad.dat"),
				[]byte("{{.HTTPHost}}"),
				[]byte(r.Host),
				-1,
			),
		)
	})
	router.GET("/proxy.pac", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.Write(pac.PacConfiguration())
	})
	return router
}
