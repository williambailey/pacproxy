package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type HTTPHandler struct {
	Pac        *Pac
	HTTPClient *http.Client
}

func NewHTTPHandler(pac *Pac) *HTTPHandler {
	return &HTTPHandler{
		Pac: pac,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DisableKeepAlives:  true,
				DisableCompression: true,
				Proxy: func(r *http.Request) (*url.URL, error) {
					return pac.Proxy(r.URL)
				},
				Dial: pac.Dial,
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return errors.New("Don't follow redirects")
			},
			Jar: nil,
		},
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) == "CONNECT" {
		h.doConnect(w, r)
	} else if r.URL.IsAbs() {
		h.doProxy(w, r)
	} else {
		h.doNonProxyRequest(w, r)
	}
}

func (h *HTTPHandler) doNonProxyRequest(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", http.StatusBadRequest)
}

func (h *HTTPHandler) doConnect(w http.ResponseWriter, r *http.Request) {
	var (
		clientConn net.Conn
		serverConn net.Conn
		err        error
	)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	clientConn, _, err = hj.Hijack()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	removeProxyHeaders(r)
	pacConn, err := h.Pac.GetPacConn(r.URL)
	if err != nil {
		http.Error(w, "", http.StatusBadGateway)
		return
	}
	if pacConn != nil {
		serverConn, err = pacConn.Dial()
		if err != nil {
			http.Error(w, "", http.StatusBadGateway)
			return
		}
		defer serverConn.Close()
		r.WriteProxy(serverConn)
	} else {
		serverConn, err = net.Dial("tcp", r.URL.Host)
		if err != nil {
			http.Error(w, "", http.StatusBadGateway)
			return
		}
		defer serverConn.Close()
		clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
	}
	defer clientConn.Close()
	defer serverConn.Close()
	go io.Copy(clientConn, serverConn)
	io.Copy(serverConn, clientConn)
}

func (h *HTTPHandler) doProxy(w http.ResponseWriter, r *http.Request) {
	removeProxyHeaders(r)
	resp, err := h.HTTPClient.Do(r)
	if err != nil {
		if resp == nil {
			http.Error(w, "", http.StatusBadGateway)
			return
		}
	}
	defer resp.Body.Close()
	wh := w.Header()
	clearHeaders(wh)
	pacResult, _ := h.Pac.CallFindProxy(r.URL)
	wh.Add("Via", fmt.Sprintf(
		"%d.%d %s (%s/%s - %s)",
		r.ProtoMajor, r.ProtoMinor,
		Name,
		Name, Version,
		pacResult,
	))
	copyHeaders(wh, resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func removeProxyHeaders(r *http.Request) {
	// this must be reset when serving a request with the client
	r.RequestURI = ""
	// If no Accept-Encoding header exists, Transport will add the headers it can accept
	// and would wrap the response body with the relevant reader.
	r.Header.Del("Accept-Encoding")
	// curl can add that, see
	// http://homepage.ntlworld.com/jonathan.deboynepollard/FGA/web-proxy-connection-header.html
	r.Header.Del("Proxy-Connection")
	// Connection is single hop Header:
	// http://www.w3.org/Protocols/rfc2616/rfc2616.txt
	// 14.10 Connection
	//   The Connection general-header field allows the sender to specify
	//   options that are desired for that particular connection and MUST NOT
	//   be communicated by proxies over further connections.
	r.Header.Del("Connection")
}

func clearHeaders(dst http.Header) {
	for k := range dst {
		dst.Del(k)
	}
}

func copyHeaders(dst, src http.Header) {
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
