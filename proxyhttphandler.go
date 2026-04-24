package main

import (
	"context"
	"crypto/tls"
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

type upstreamAuth struct {
	user  string
	pass  string
	hosts []string
}

func (a *upstreamAuth) Header() string {
	raw := a.user + ":" + a.pass
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(raw))
}

// shouldApply reports whether the upstream auth credentials should be sent
// to the given proxy. Entries can be "host" (any port) or "host:port"
// (exact match). An empty host list means "never send credentials".
func (a *upstreamAuth) shouldApply(p pac.Proxy) bool {
	if a == nil || len(a.hosts) == 0 {
		return false
	}
	host := strings.ToLower(p.Hostname)
	hostPort := fmt.Sprintf("%s:%d", host, p.Port)
	for _, h := range a.hosts {
		if h == host || h == hostPort {
			return true
		}
	}
	return false
}

// shouldApplyURL is the URL-keyed equivalent of shouldApply, used when we
// only have the resolved proxy URL on hand (CONNECT path).
func (a *upstreamAuth) shouldApplyURL(u *url.URL) bool {
	if a == nil || u == nil || len(a.hosts) == 0 {
		return false
	}
	host := strings.ToLower(u.Hostname())
	hostPort := strings.ToLower(u.Host)
	for _, h := range a.hosts {
		if h == host || h == hostPort {
			return true
		}
	}
	return false
}

type proxyHTTPHandler struct {
	proxyFinder     pac.ProxyFinder
	proxySelector   pac.ProxySelector
	httpClient      *http.Client
	dialer          *net.Dialer
	resolver        *customResolver
	nonProxyHandler http.Handler
	upstreamAuth    *upstreamAuth
}

func newProxyHTTPHandler(
	proxyFinder pac.ProxyFinder,
	proxySelector pac.ProxySelector,
	nonProxyHandler http.Handler,
	upstreamAuth *upstreamAuth,
	resolver *customResolver,
) *proxyHTTPHandler {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
	}
	dialContext := dialer.DialContext
	if resolver != nil {
		dialContext = resolver.dialContext(dialer)
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
		DialContext:           dialContext,
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
		resolver:        resolver,
		nonProxyHandler: nonProxyHandler,
		upstreamAuth:    upstreamAuth,
	}
	transport.Proxy = handler.lookupProxy
	return handler
}

func (h *proxyHTTPHandler) resolveAddr(ctx context.Context, hostPort string) (string, error) {
	if h.resolver == nil {
		return hostPort, nil
	}
	return h.resolver.resolveHost(ctx, hostPort)
}

func (h *proxyHTTPHandler) dialTCP(ctx context.Context, hostPort string) (net.Conn, error) {
	addr, err := h.resolveAddr(ctx, hostPort)
	if err != nil {
		return nil, err
	}
	return h.dialer.DialContext(ctx, "tcp", addr)
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
	proxies, err := h.proxyFinder.FindProxyForURL(r.URL)
	if err != nil {
		return nil, err
	}
	proxy := h.proxySelector.SelectProxy(proxies)
	log.Printf("Proxy Lookup %q, got %q. Selected %q", r.URL, proxies, proxy)
	if proxy == pac.DirectProxy {
		return nil, nil
	}
	scheme := "http"
	if proxy.TLS {
		scheme = "https"
	}
	proxyURL := &url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", proxy.Hostname, proxy.Port),
	}
	if proxyAuth := r.Header.Get("Proxy-Authorization"); proxyAuth != "" {
		if u, p, ok := parseBasicAuth(proxyAuth); ok {
			proxyURL.User = url.UserPassword(u, p)
		}
	} else if h.upstreamAuth.shouldApply(proxy) {
		proxyURL.User = url.UserPassword(h.upstreamAuth.user, h.upstreamAuth.pass)
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
		serverConn, err = h.dialTCP(r.Context(), r.URL.Host)
		if err != nil {
			log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer serverConn.Close()
	} else {
		hostPort := proxyURL.Hostname() + ":" + proxyURL.Port()
		resolved, rerr := h.resolveAddr(r.Context(), hostPort)
		if rerr != nil {
			log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, rerr)
			http.Error(w, rerr.Error(), http.StatusBadGateway)
			return
		}
		if proxyURL.Scheme == "https" {
			serverConn, err = tls.DialWithDialer(h.dialer, "tcp", resolved, &tls.Config{
				ServerName: proxyURL.Hostname(),
			})
		} else {
			serverConn, err = h.dialer.Dial("tcp", resolved)
		}
		if err != nil {
			log.Printf("HTTP Connect Proxy %q: %d %s", r.URL, http.StatusBadGateway, err)
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer serverConn.Close()
		removeProxyHeaders(r)
		if h.upstreamAuth.shouldApplyURL(proxyURL) && r.Header.Get("Proxy-Authorization") == "" {
			r.Header.Set("Proxy-Authorization", h.upstreamAuth.Header())
		}
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
		serverConn.SetDeadline(time.Now().Add(10 * time.Millisecond))
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
