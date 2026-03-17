package main

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/williambailey/pacproxy/pac"
)

type directProxyFinder struct{}

func (f *directProxyFinder) FindProxyForURL(in *url.URL) (pac.Proxies, error) {
	return pac.Proxies{pac.DirectProxy}, nil
}

type fixedProxyFinder struct {
	proxy pac.Proxy
}

func (f *fixedProxyFinder) FindProxyForURL(in *url.URL) (pac.Proxies, error) {
	return pac.Proxies{f.proxy}, nil
}

func newTestHandler(finder pac.ProxyFinder) *proxyHTTPHandler {
	return newProxyHTTPHandler(finder, &pac.FirstItemSelector{}, nil)
}

func TestConnectDirectClosesFDs(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer target.Close()

	handler := newTestHandler(&directProxyFinder{})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	proxyURL, _ := url.Parse(proxy.URL)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	client.CloseIdleConnections()

	if string(body) != "hello" {
		t.Errorf("expected body %q, got %q", "hello", string(body))
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestConnectViaUpstreamProxyClosesFDs(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("proxied"))
	}))
	defer target.Close()

	upstream := httptest.NewServer(newTestHandler(&directProxyFinder{}))
	defer upstream.Close()

	upstreamHost, upstreamPort, _ := net.SplitHostPort(upstream.Listener.Addr().String())
	handler := newTestHandler(&fixedProxyFinder{
		proxy: pac.Proxy{Hostname: upstreamHost, Port: mustAtoi(upstreamPort)},
	})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	proxyURL, _ := url.Parse(proxy.URL)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	client.CloseIdleConnections()

	if string(body) != "proxied" {
		t.Errorf("expected body %q, got %q", "proxied", string(body))
	}
}

func TestConnectTunnelClosesOnServerDisconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverClosed := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close()
		close(serverClosed)
	}()

	handler := newTestHandler(&directProxyFinder{})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	conn, err := net.DialTimeout("tcp", proxy.Listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	connectReq := "CONNECT " + listener.Addr().String() + " HTTP/1.1\r\nHost: " + listener.Addr().String() + "\r\n\r\n"
	conn.Write([]byte(connectReq))

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	response := string(buf[:n])
	if response[:12] != "HTTP/1.0 200" {
		t.Fatalf("expected 200, got %q", response)
	}

	select {
	case <-serverClosed:
	case <-time.After(2 * time.Second):
		t.Fatal("server accept timed out")
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buf)
	if err == nil {
		t.Error("expected connection to be closed after server disconnect")
	}
}

func TestConnectTunnelClosesOnClientDisconnect(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	serverConnClosed := make(chan struct{})
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		io.Copy(io.Discard, conn)
		close(serverConnClosed)
	}()

	handler := newTestHandler(&directProxyFinder{})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	conn, err := net.DialTimeout("tcp", proxy.Listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}

	connectReq := "CONNECT " + listener.Addr().String() + " HTTP/1.1\r\nHost: " + listener.Addr().String() + "\r\n\r\n"
	conn.Write([]byte(connectReq))

	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	response := string(buf[:n])
	if response[:12] != "HTTP/1.0 200" {
		t.Fatalf("expected 200, got %q", response)
	}

	conn.Close()

	select {
	case <-serverConnClosed:
	case <-time.After(2 * time.Second):
		t.Fatal("server-side connection was not closed after client disconnect")
	}
}

func TestHTTPProxyForward(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("forwarded"))
	}))
	defer target.Close()

	handler := newTestHandler(&directProxyFinder{})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	proxyURL, _ := url.Parse(proxy.URL)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(target.URL)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "forwarded" {
		t.Errorf("expected body %q, got %q", "forwarded", string(body))
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPProxyStripsHopByHopHeaders(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Proxy-Connection") != "" {
			t.Error("Proxy-Connection header was not stripped")
		}
		if r.Header.Get("Connection") != "" {
			t.Error("Connection header was not stripped")
		}
		w.Write([]byte("ok"))
	}))
	defer target.Close()

	handler := newTestHandler(&directProxyFinder{})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	req, _ := http.NewRequest("GET", target.URL, nil)
	req.Header.Set("Proxy-Connection", "keep-alive")
	req.Header.Set("Connection", "keep-alive")

	proxyURL, _ := url.Parse(proxy.URL)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()
}

func TestConnectDialFailureReturns502(t *testing.T) {
	handler := newTestHandler(&directProxyFinder{})
	proxy := httptest.NewServer(handler)
	defer proxy.Close()

	conn, err := net.DialTimeout("tcp", proxy.Listener.Addr().String(), 2*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	connectReq := "CONNECT 127.0.0.1:1 HTTP/1.1\r\nHost: 127.0.0.1:1\r\n\r\n"
	conn.Write([]byte(connectReq))

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	response := string(buf[:n])
	if response[:12] != "HTTP/1.1 502" {
		t.Errorf("expected 502, got %q", response)
	}
}

func mustAtoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
