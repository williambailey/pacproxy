package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const Name = "pacproxy"
const Version = "0.8.0"

var (
	fPac     string
	fListen  string
	fVerbose bool
)

func init() {
	flag.StringVar(&fPac, "c", "proxy.pac", "PAC file to use")
	flag.StringVar(&fListen, "l", "127.0.0.1:12345", "Interface and port to listen on")
	flag.BoolVar(&fVerbose, "v", false, "send verbose output to STDERR")
}

func main() {
	flag.Parse()
	if fVerbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}
	log.SetPrefix("")
	log.SetFlags(log.Ldate | log.Lmicroseconds)

	pac, err := NewPac()
	if err != nil {
		log.Fatal(err)
	}
	err = pac.LoadFile(fPac)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on %q", fListen)
	log.Fatal(
		http.ListenAndServe(
			fListen,
			NewProxyHTTPHandler(pac, NewNonProxyHTTPHandler(pac)),
		),
	)
}
