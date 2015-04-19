package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
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
const Version = "0.8.0"

var (
	fPac           string
	fListen        string
	fVerbose       bool
	pac            gopac.Parser
	pacRecordSplit *regexp.Regexp
	pacItemSplit   *regexp.Regexp
)

func init() {
	pacRecordSplit = regexp.MustCompile(`\s*;\s*`)
	pacItemSplit = regexp.MustCompile(`\s+`)
	flag.StringVar(&fPac, "c", "proxy.pac", "PAC file to use")
	flag.StringVar(&fListen, "l", "127.0.0.1:12345", "Interface and port to listen on")
	flag.BoolVar(&fVerbose, "v", false, "send verbose output to STDERR")
}

func main() {
	flag.Parse()
	logWriter := ioutil.Discard
	if fVerbose {
		logWriter = os.Stderr
	}
	log.SetOutput(logWriter)
	log.SetPrefix("")
	log.SetFlags(log.Ldate | log.Lmicroseconds)
	log := log.New(logWriter, "", log.Flags())
	pacLookup := &pacLookup{
		pac: &gopac.Parser{},
		log: log,
	}
	err := pacLookup.pac.Parse(fPac)
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:  true,
			DisableCompression: true,
			Proxy: func(r *http.Request) (*url.URL, error) {
				p, err := pacLookup.fetchOne(r.URL)
				if err != nil {
					log.Printf("Failed to get proxy configuration from pac: %s", err)
					return nil, err
				}
				if p != nil {
					log.Printf("Using proxy %v", p)
				} else {
					log.Printf("Going direct")
				}
				return p, err
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return errors.New("Don't follow redirects")
		},
	}
	handler := &httpHandler{
		client:    httpClient,
		pacLookup: pacLookup,
		log:       log,
	}
	log.Printf("Listening on %q", fListen)
	log.Fatal(http.ListenAndServe(fListen, handler))
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
	if o := strings.Index(u.Host, ":"); o >= 0 {
		pacResult, err = l.pac.FindProxy(u.String(), u.Host[:o])
	} else {
		pacResult, err = l.pac.FindProxy(u.String(), u.Host)
	}
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
		case "SOCKS":
			return nil, errors.New("SOCKS is not supported")
		default:
			return nil, fmt.Errorf("Unknown PAC command %q", p[0])
		}
	}
	if len(r) == 0 {
		r = append(r, nil)
	}
	return r, nil
}

func (l *pacLookup) fetchOne(u *url.URL) (*url.URL, error) {
	results, err := l.fetch(u)
	if err != nil {
		return nil, err
	}
	for _, proxyURL := range results {
		// TODO: failover proxy support.
		return proxyURL, nil
	}
	return nil, nil
}

type httpHandler struct {
	client    *http.Client
	log       *log.Logger
	pacLookup *pacLookup
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.log.Printf("Got request %s %s", r.Method, r.URL)
	if strings.ToUpper(r.Method) == "CONNECT" {
		h.doConnect(w, r)
		return
	}
	if !r.URL.IsAbs() {
		h.doNonProxyRequest(w, r)
		return
	}
	h.doProxy(w, r)
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
	proxyURL, err = h.pacLookup.fetchOne(r.URL)
	if err != nil {
		h.log.Printf("Failed to get proxy configuration from pac: %s", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
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
	resp, err := h.client.Do(r)
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
