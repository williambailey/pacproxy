package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func NewNonProxyHTTPHandler(pac *Pac) http.Handler {
	router := httprouter.New()
	router.RedirectFixedPath = false
	router.RedirectTrailingSlash = false
	router.Handler("GET", "/pacproxy/*filepath", http.StripPrefix("/pacproxy/", http.FileServer(assetFS())))
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "%s v%s %s\n", Name, Version, pac.PacFilename())
	})
	router.GET("/wpad.dat", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		wpad := strings.Replace(`
function FindProxyForURL(url, host)
{
	if (isInNet(host, "127.0.0.0", "255.0.0.0"))
	{
		return "DIRECT";
	}
	return "PROXY %%HOST%%";
}
`, "%%HOST%%", r.Host, -1)
		w.Write([]byte(wpad))
	})
	router.GET("/proxy.pac", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
		w.Write(pac.PacConfiguration())
	})
	return router
}
