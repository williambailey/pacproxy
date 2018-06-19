package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/williambailey/pacproxy/pac"
	"net"
)

// Name of the app
const Name = "pacproxy"

// Version of the app
const Version = "2.0.0"

var (
	fPac          string
	fListen       string
	fListenSocks5 string
	fVerbose      bool
)

func init() {
	flag.StringVar(&fPac, "c", "", "PAC file name, url or javascript to use")
	flag.StringVar(&fListen, "l", "127.0.0.1:8080", "Interface and port to listen on of http proxy")
	flag.StringVar(&fListenSocks5, "s", "127.0.0.1:1080", "Interface and port to listen on of socks5 proxy")
	flag.BoolVar(&fVerbose, "v", false, "send verbose output to STDERR")
}
func serveHttpProxy(proxyFinder pac.ProxyFinder, proxySelector pac.ProxySelector) {
	srv := &http.Server{
		Addr:              fListen,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       60 * time.Second,
		Handler: newProxyHTTPHandler(
			proxyFinder,
			proxySelector,
			newNonProxyHTTPHandler(),
		),
	}
	log.Printf("Http Proxy listening on %q", fListen)
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
func serveSocks5Proxy(proxyFinder pac.ProxyFinder, proxySelector pac.ProxySelector) {
	log.Printf("Socks5 Proxy listening on %q", fListen)
	l, err := net.Listen("tcp", fListenSocks5)
	if err != nil {
		log.Panic(err)
	}
	handler := newProxySocksHandler(proxyFinder, proxySelector)
	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handler.Handle(client)
	}
}
func main() {
	flag.Parse()
	if fVerbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	log.SetPrefix("")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("Starting %s v%s", Name, Version)

	otto := pac.NewOttoEngine(
		pac.OttoLoader(pac.SmartLoader(fPac)),
	)
	if err := otto.Start(); err != nil {
		log.Panic(err)
	}
	defer otto.Stop()

	initSignalNotify(otto)
	finish := make(chan bool)
	selector := &pac.FirstItemSelector{}
	go func() {
		serveHttpProxy(otto, selector)
		finish <- true
	}()
	go func() {
		serveSocks5Proxy(otto, selector)
		finish <- true
	}()
	<-finish
}
