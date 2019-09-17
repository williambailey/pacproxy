package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"fmt"

	"github.com/williambailey/pacproxy/pac"
)

// Name of the app
const Name = "pacproxy"

// Version of the app
const Version = "2.0.2"

var (
	fPac     string
	fListen  string
	fVerbose bool
)

func init() {
	flag.StringVar(&fPac, "c", "", "PAC file name, url or javascript to use (required)")
	flag.StringVar(&fListen, "l", "127.0.0.1:8080", "Interface and port to listen on")
	flag.BoolVar(&fVerbose, "v", false, "send verbose output to STDERR")
}

func main() {
	required := []string{"c"}
      	flag.Parse()
	
	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
        for _, req := range required {
	    if !seen[req] {
	        fmt.Fprintf(os.Stderr, "missing required -%s argument/flag\n", req)
	        os.Exit(2) // the same exit code flag.Parse uses
	    }
        }
	
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

	srv := &http.Server{
		Addr:              fListen,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       60 * time.Second,
		Handler: newProxyHTTPHandler(
			otto,
			&pac.FirstItemSelector{},
			newNonProxyHTTPHandler(),
		),
	}
	log.Printf("Listening on %q", fListen)
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
