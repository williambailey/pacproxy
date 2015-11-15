package main

//go:generate go-bindata-assetfs -pkg $GOPACKAGE -nomemcopy -nocompress -o bindata.go -prefix "resource/bindata/" resource/bindata/...
//go:generate gofmt -w bindata_assetfs.go

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const Name = "pacproxy"
const Version = "1.0.0"

var (
	fPac     string
	fListen  string
	fVerbose bool
)

func init() {
	flag.StringVar(&fPac, "c", "", "PAC file to use")
	flag.StringVar(&fListen, "l", "127.0.0.1:8080", "Interface and port to listen on")
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
	log.Printf("Starting %s v%s", Name, Version)

	pac, err := NewPac()
	if err != nil {
		log.Fatal(err)
	}
	if fPac != "" {
		err = pac.LoadFile(fPac)
		if err != nil {
			log.Fatal(err)
		}
	}

	initSignalNotify(pac)

	log.Printf("Listening on %q", fListen)
	log.Fatal(
		http.ListenAndServe(
			fListen,
			NewProxyHTTPHandler(pac, NewNonProxyHTTPHandler(pac)),
		),
	)
}
