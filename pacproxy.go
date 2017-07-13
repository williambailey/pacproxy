package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/williambailey/pacproxy/pac"
)

// Name of the app
const Name = "pacproxy"

// Version of the app
const Version = "2.0.0-beta.1"

var (
	fPac     string
	fListen  string
	fVerbose bool
)

func init() {
	flag.StringVar(&fPac, "c", "", "PAC file name, url or javascript to use")
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
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile | log.LUTC)
	log.Printf("Starting %s v%s", Name, Version)

	for {
		if err := listenAndServe(fPac, fListen); err != nil {
			log.Panic(err)
		}
	}
}

func listenAndServe(pacFile, listen string) error {
	defer func() {
		if r := recover(); r != nil {
			log.Print("Recovered")
		}
	}()
	otto := pac.NewOttoEngine(
		pac.OttoLoader(pac.SmartLoader(pacFile)),
	)
	if err := otto.Start(); err != nil {
		return err
	}
	defer otto.Stop()

	initSignalNotify(otto)

	srv := &http.Server{
		Addr:              listen,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       60 * time.Second,
		Handler: newProxyHTTPHandler(
			otto,
			&pac.FirstItemSelector{},
			newNonProxyHTTPHandler(),
		),
	}
	log.Printf("Listening on %q", listen)
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
	return nil
}
