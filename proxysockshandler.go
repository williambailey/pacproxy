package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"time"

	"bufio"
	"github.com/williambailey/pacproxy/pac"
	"net/http"
	"strconv"
)

type proxySocks5Handler struct {
	proxyFinder   pac.ProxyFinder
	proxySelector pac.ProxySelector
	dialer        *net.Dialer
}

func newProxySocksHandler(
	proxyFinder pac.ProxyFinder,
	proxySelector pac.ProxySelector,
) *proxySocks5Handler {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 10 * time.Second,
	}
	handler := &proxySocks5Handler{
		proxyFinder,
		proxySelector,
		dialer,
	}
	return handler
}

func (h *proxySocks5Handler) lookupProxy(targetUrl *url.URL) (string, error) {
	proxies, err := h.proxyFinder.FindProxyForURL(targetUrl)
	if err != nil {
		return "", err
	}
	proxy := h.proxySelector.SelectProxy(proxies)
	log.Printf("Proxy Lookup %q, got %q. Selected %q", targetUrl, proxies, proxy)
	if proxy == pac.DirectProxy {
		return "", nil
	}
	host := fmt.Sprintf("%s:%d", proxy.Hostname, proxy.Port)
	return host, nil
}
func (h *proxySocks5Handler) Handle(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	var b [1024]byte
	n, err := client.Read(b[:])
	if err != nil {
		log.Printf("Socks5 Proxy read header failed: %s", err)
		return
	}

	// only handle socks5 protocol
	if b[0] != 0x05 {
		return
	}

	// tell clients no need to auth
	client.Write([]byte{0x05, 0x00})

	//get target host port
	n, err = client.Read(b[:])
	var host, port string
	switch b[3] {
	case 0x01: // ipv4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: // domain
		host = string(b[5 : n-2]) //b[4] length of domain
	case 0x04: // ipv6
		host = net.IP(b[4:19]).String()
	}
	port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))

	backendAddr := net.JoinHostPort(host, port)
	backendUrl, err := url.Parse("http://" + backendAddr)
	if err != nil {
		log.Printf("Socks5 Proxy parse url failed: %s", err)
		return
	}
	// get proxy
	proxyAddr, err := h.lookupProxy(backendUrl)
	if err != nil {
		log.Printf("Socks5 Proxy find proxy failed: %s", err)
		return
	}
	// dial with CONNECT to http proxy
	server, err := DialHttpProxy(proxyAddr, backendAddr, nil)
	if err != nil {
		log.Printf("Socks5 Proxy connect proxy failed: %s", err)
		return
	}
	// tell client connection ready
	client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

	// pipe client and backend
	go io.Copy(server, client)
	io.Copy(client, server)
}

func DialHttpProxy(proxyAddr, backend string, auth *url.Userinfo) (net.Conn, error) {
	// reference fellows:
	// https://github.com/mwitkow/go-http-dialer
	// https://github.com/golang/go/issues/17227
	// https://gist.github.com/jim3ma/3750675f141669ac4702bc9deaf31c6b
	// https://www.jianshu.com/p/172810a70fad
	req := &http.Request{
		Method: "CONNECT",
		URL: &url.URL{
			Host: backend,
		},
		Host:   backend,
		Header: make(http.Header),
		Close:  false,
	}
	if auth != nil {
		password, _ := auth.Password()
		req.SetBasicAuth(auth.Username(), password)
	}
	req.Header.Set("Proxy-Connection", "Keep-Alive")

	conn, err := net.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, err
	}

	err = req.Write(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		conn.Close()
		return nil, err
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		conn.Close()
		return nil, fmt.Errorf("http proxy status=%d", resp.StatusCode)
	}
	return conn, nil
}
