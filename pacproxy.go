package main

//go:generate go-bindata-assetfs -pkg $GOPACKAGE -nomemcopy -nocompress -o bindata.go -prefix "resource/bindata/" resource/bindata/...

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGUSR1)
	go func() {
		for s := range sigChan {
			switch s {
			case syscall.SIGHUP:
				f := pac.PacFilename()
				if f == "" {
					log.Println("Cleaning connection statuses however the current PAC configuration was not loaded from a file.")
					pac.ConnService.Clear()
					return
				}
				log.Printf("Cleaning connection statuses and reloading PAC configuration from %q.\n", f)
				if e := pac.LoadFile(f); e != nil {
					log.Println(e)
				}
			case syscall.SIGUSR1:
				knownProxies := pac.ConnService.KnownProxies()
				log.Printf("Known proxies: %d\n", len(knownProxies))
				var (
					s string
					i int
				)
				for _, p := range knownProxies {
					i++
					s = fmt.Sprintf("%3d. %s - ", i, p.Address())
					bl := p.BlacklistDuration()
					if bl == 0 {
						s += fmt.Sprint("Active")
					} else {
						s += fmt.Sprintf("Blacklisted (%s)", bl)
					}
					log.Println(s)
				}
			}
		}
	}()

	log.Printf("Listening on %q", fListen)
	log.Fatal(
		http.ListenAndServe(
			fListen,
			NewProxyHTTPHandler(pac, NewNonProxyHTTPHandler(pac)),
		),
	)
}
