// +build !plan9

package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/williambailey/pacproxy/pac"
)

func initSignalNotify(pac pac.EngineManager) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	go func() {
		for s := range sigChan {
			switch s {
			case syscall.SIGHUP:
				log.Print("SIGHUP")
				if err := pac.Reload(); err != nil {
					log.Panic(err)
				}
			}
		}
	}()

}
