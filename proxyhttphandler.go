package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/williambailey/pacproxy/pac"
)

type proxyHTTPHandler struct {
	proxyFinder     pac.ProxyFinder
	proxySelector   pac.ProxySelector
	httpClient      *http.Client
	dialer          *net.Dialer
	nonProxyHandler http.Handler
}

func newProxyHTTPHandler(
	proxyFinder pac.ProxyFinder,
	proxySelector pac.ProxySelector,
	nonProxyHandler http.Handler,
) *proxyHTTPHandler {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
	}
	transport := &http.Transport{
		DisableKeepAlives:     false,
		DisableCompression:    false,
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       5 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		Proxy:                 nil,
		Dial:                  dialer.Dial,
	}
	handler := &proxyHTTPHandler{
		proxyFinder:   proxyFinder,
		proxySelector: proxySelector,
		httpClient: &http.Client{
			Transport: transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Don't follow redirects, but do return their contents.
				return http.ErrUseLastResponse
			},
			Jar: nil,
		},
		dialer:          dialer,
		nonProxyHandler: nonProxyHandler,
	}
	transport.Proxy = handler.lookupProxy
	return handler
}

func (h *proxyHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.ToUpper(r.Method) == "CONNECT" {
		h.doConnectProxy(w, r)
	} else if r.URL.IsAbs() {
		h.doHTTPProxy(w, r)
	} else if h.nonProxyHandler != nil {
		h.nonProxyHandler.ServeHTTP(w, r)
	} else {
		http.Error(w, "", http.StatusBadRequest)
	}
}

func (h *proxyHTTPHandler) lookupProxy(r *http.Request) (*url.URL, error) {
	sourceURL := &url.URL{
		Scheme:      fScheme,
		Opaque:      r.URL.Opaque,
		User:        r.URL.User,
		Host:        r.URL.Host,
		Path:        r.URL.Path,
		RawPath:     r.URL.RawPath,
		ForceQuery:  r.URL.ForceQuery,
		RawQuery:    r.URL.RawQuery,
		Fragment:    r.URL.Fragment,
		RawFragment: r.URL.RawFragment,
	}
	proxies, err := h.proxyFinder.FindProxyForURL(sourceURL)
	if err != nil {
		return nil, err
	}
	proxy := h.proxySelector.SelectProxy(proxies)
	log.Printf("Proxy Lookup %q, got %q. Selected %q", sourceURL, proxies, proxy)
	if proxy == pac.DirectProxy {
		return nil, nil
	}
	proxyURL := &url.URL{
		Host: fmt.Sprintf("%s:%d", proxy.Hostname, proxy.Port),
	}
	if proxyAuth := r.Header.Get("Proxy-Authorization"); proxyAuth != "" {
		if u, p, ok := parseBasicAuth(proxyAuth); ok {
			proxyURL.User = url.UserPassword(u, p)
		}
	}
	return proxyURL, nil
}

func (h *proxyHTTPHandler) doConnectProxy(w http.ResponseWriter, r *http.Request) {
	var (
		clientConn net.Conn
		serverConn net.Conn
		err        error
	)

	proxyURL, err := h.lookupProxy(r)
	if err != nil {
		log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	if proxyURL == nil {
		serverConn, err = h.dialer.Dial("tcp", r.URL.Host)
		if err != nil {
			log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer serverConn.Close()
	} else {
		serverConn, err = h.dialer.Dial("tcp", proxyURL.Hostname()+":"+proxyURL.Port())
		if err != nil {
			log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer serverConn.Close()
		removeProxyHeaders(r)
		//r.WriteProxy(serverConn)
		r.Write(serverConn) // instead of WriteProxy as this will *hopefully* deal with CONNECT correctly.
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		err = errors.New("unable to get hijacker")
		log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	clientConn, _, err = hj.Hijack()
	if err != nil {
		log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer clientConn.Close()
	if proxyURL == nil {
		clientConn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(clientConn, serverConn)
		clientConn.SetDeadline(time.Now().Add(10 * time.Millisecond))
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		io.Copy(serverConn, clientConn)
	}()
	wg.Wait()
}

func (h *proxyHTTPHandler) doHTTPProxy(w http.ResponseWriter, r *http.Request) {
	removeProxyHeaders(r)
	resp, err := h.httpClient.Do(r)
	if err != nil && resp == nil {
		log.Printf("HTTP Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	wh := w.Header()
	clearHeaders(wh)
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

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}
