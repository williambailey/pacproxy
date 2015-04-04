package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/jackwakefield/gopac"
)

const Name = "pacproxy"
const Version = "0.1.0"

var (
	pac               gopac.Parser
	pacRecordSplit    *regexp.Regexp
	pacItemSplit      *regexp.Regexp
	errClientRedirect error
)

func init() {
	pacRecordSplit = regexp.MustCompile(`\s*;\s*`)
	pacItemSplit = regexp.MustCompile(`\s+`)
	errClientRedirect = errors.New("Don't follow redirects.")
}

func main() {
	log := log.New(os.Stderr, "", log.LstdFlags)
	pacLookup := &pacLookup{
		pac: &gopac.Parser{},
		log: log,
	}
	err := pacLookup.pac.Parse("proxy.pac")
	if err != nil {
		log.Fatal(err)
	}

	handler := &httpHandler{
		pacLookup: pacLookup,
		log:       log,
	}
	log.Fatal(http.ListenAndServe("127.0.0.1:12345", handler))

}

type pacLookup struct {
	pac *gopac.Parser
	log *log.Logger
}

func (l *pacLookup) fetchString(u *url.URL) string {
	str, err := l.pac.FindProxy(u.String(), u.Host)
	if err != nil {
		return ""
	}
	return str
}

func (l *pacLookup) fetch(u *url.URL) ([]*url.URL, error) {
	var (
		err       error
		pacResult string
		proxyURL  *url.URL
	)
	r := make([]*url.URL, 0, 10)
	pacResult, err = l.pac.FindProxy(u.String(), u.Host)
	if err != nil {
		return nil, err
	}
	for _, rSplit := range pacRecordSplit.Split(pacResult, 10) {
		p := pacItemSplit.Split(rSplit, 2)
		switch strings.ToUpper(p[0]) {
		case "DIRECT":
			r = append(r, nil)
		case "PROXY":
			proxyURL, err = url.Parse("http://" + p[1])
			if err != nil {
				return nil, err
			}
			r = append(r, proxyURL)
		}
	}
	if len(r) == 0 {
		r = append(r, nil)
	}
	return r, nil
}

type httpHandler struct {
	log       *log.Logger
	pacLookup *pacLookup
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Printf("Got request %s %s", r.Method, r.URL)
	if r.Method == "CONNECT" {
		h.doConnect(w, r)
		return
	}
	if !r.URL.IsAbs() {
		h.doNonProxyRequest(w, r)
		return
	}
	h.doProxy(w, r)
}

func (h *httpHandler) client() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:  true,
			DisableCompression: true,
			Proxy: func(r *http.Request) (*url.URL, error) {
				pacResults, err := h.pacLookup.fetch(r.URL)
				if err != nil {
					h.log.Printf("Failed to get proxy configuration from pac: %s", err)
					return nil, err
				}
				for _, proxyURL := range pacResults {
					// TODO: failover proxy support.
					h.log.Printf("Using proxy %v", proxyURL)
					return proxyURL, nil
				}
				h.log.Print("Direct connection")
				return nil, nil
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errClientRedirect
		},
	}
}

func (h *httpHandler) doConnect(w http.ResponseWriter, r *http.Request) {
	var (
		clientConn net.Conn
		serverConn net.Conn
		err        error
		proxyURL   *url.URL
	)

	hj, ok := w.(http.Hijacker)
	if !ok {
		h.log.Print("Unable to hijack connection")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	clientConn, _, err = hj.Hijack()
	if err != nil {
		h.log.Printf("Failed to hijack connection: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	removeProxyHeaders(r)
	pacResults, err := h.pacLookup.fetch(r.URL)
	if err != nil {
		h.log.Printf("Failed to get proxy configuration from pac: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	for _, proxyURL = range pacResults {
		// TODO: failover proxy support.
		break
	}
	if proxyURL != nil {
		h.log.Printf("Using proxy connect %v", proxyURL)
		h.log.Printf("Dial %v", proxyURL.Host)
		serverConn, err = net.Dial("tcp", proxyURL.Host)
		if err != nil {
			h.log.Printf("Failed to dial: %s", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer serverConn.Close()
		r.WriteProxy(serverConn)
	} else {
		h.log.Print("Using direct connect")
		h.log.Printf("Dial %v", r.URL.Host)
		serverConn, err = net.Dial("tcp", r.URL.Host)
		if err != nil {
			h.log.Printf("Failed to dial: %s", err)
			http.Error(w, "", http.StatusInternalServerError)
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

func (h *httpHandler) doProxy(w http.ResponseWriter, r *http.Request) {
	removeProxyHeaders(r)
	client := h.client()
	resp, err := client.Do(r)
	if err != nil {
		if resp == nil {
			h.log.Printf("Client do err: %s", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
	wh := w.Header()
	clearHeaders(wh)
	wh.Add("Via", fmt.Sprintf("%d.%d %s (%s/%s - %s)", r.ProtoMajor, r.ProtoMinor, Name, Name, Version, h.pacLookup.fetchString(r.URL)))
	copyHeaders(wh, resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func (h *httpHandler) doNonProxyRequest(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "This is a proxy server. Does not respond to non-proxy requests.", http.StatusBadRequest)
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
