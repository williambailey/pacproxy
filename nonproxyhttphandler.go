package main

import (
	"bytes"
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
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
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

	router.GET("/stats", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "{\n")
		first := true
		expvar.Do(func(kv expvar.KeyValue) {
			if !first {
				fmt.Fprintf(w, ",\n")
			}
			first = false
			fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
		})
		fmt.Fprintf(w, "\n}\n")
	})

	router.GET("/status", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

		HTMLTemplate.ExecuteTemplate(w, "status", d)
	})
	router.GET("/wpad.dat", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig; charset=utf-8")
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
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig; charset=utf-8")
		w.Write(pac.PacConfiguration())
	})
	router.GET("/pac/find-proxy", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		urlParam := r.URL.Query().Get("url")
		if urlParam == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		urlURL, err := url.Parse(urlParam)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		pacResult, err := pac.CallFindProxyForURLFromURL(urlURL)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(pacResult))
	})
	return router
}
