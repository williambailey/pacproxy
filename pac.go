package main

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"sync"
)

const pacDefaultJavascript = `
function FindProxyForURL(url, host)
{
	return "DIRECT";
}
`

// Pac is the main proxy auto configuration engine.
type Pac struct {
	mutex   *sync.Mutex
	runtime *gopacRuntime
}

// NewPac create a new pac instance.
func NewPac() (*Pac, error) {
	p := &Pac{
		mutex: &sync.Mutex{},
	}
	if err := p.Load(pacDefaultJavascript); err != nil {
		return nil, err
	}
	return p, nil
}

// Unload any previously loaded pac configuration and referts to default.
func (p *Pac) Unload() error {
	return p.Load(pacDefaultJavascript)
}

// Load attempts to load a pac from a string, a byte slice,
// a bytes.Buffer, or an io.Reader, but it MUST always be in UTF-8.
func (p *Pac) Load(js interface{}) error {
	var err error
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.runtime, err = newGopacRuntime()
	if err != nil {
		return err
	}
	if _, err := p.runtime.vm.Run(js); err != nil {
		return err
	}
	return nil
}

// LoadFile attempt to load a pac file.
func (p *Pac) LoadFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	return p.Load(f)
}

// CallFindProxyForURL using the current pac.
func (p *Pac) CallFindProxyForURL(url, host string) (string, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.runtime.findProxyForURL(url, host)
}

// FindProxy return that URL of the proxy that the client should use.
// If the client should establish a direct connect that it will return
// nil.
func (p *Pac) FindProxy(in *url.URL) (u *url.URL, err error) {
	if in == nil {
		return
	}
	var (
		h string
		s string
	)
	if o := strings.Index(in.Host, ":"); o >= 0 {
		h = in.Host[:o]
	} else {
		h = in.Host
	}
	if s, err = p.CallFindProxyForURL(in.String(), h); err != nil {
		return
	}
	if s == "DIRECT" {
		return
	}
	err = errors.New("Not implemented.")
	return
}
