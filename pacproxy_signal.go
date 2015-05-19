// +build !plan9

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func initSignalNotify(pac *Pac) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
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
			}
		}
	}()

}
