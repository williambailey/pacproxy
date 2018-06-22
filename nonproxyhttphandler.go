package main

//go:generate make favicon

import (
	"fmt"
	"net/http"
)

func newNonProxyHTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			w.Write(faviconIco)
			return
		}
		http.Error(
			w,
			fmt.Sprintf("%s %s\nhttps://github.com/shakirshakiel/pacproxy", Name, Version),
			http.StatusBadGateway,
		)
	})
}
