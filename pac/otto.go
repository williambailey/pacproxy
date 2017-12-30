package pac

import (
	"errors"
	"log"
	"net/url"
	"sync"

	"github.com/robertkrimen/otto"
	"github.com/williambailey/pacproxy/pacfunc"
)

// OttoEngineOpt used to configure an OttoEngine via the NewOttoEngine func
type OttoEngineOpt func(*OttoEngine)

// OttoLoader that the engine should use
func OttoLoader(fn Loader) OttoEngineOpt {
	return func(o *OttoEngine) {
		o.loader = fn
	}
}

// OttoStringLoader implements a string loader
func OttoStringLoader(pac string) OttoEngineOpt {
	return OttoLoader(func() (string, error) {
		return pac, nil
	})
}

// NewOttoEngine instance with configuration
func NewOttoEngine(opts ...OttoEngineOpt) *OttoEngine {
	otto := &OttoEngine{
		mutex: &sync.RWMutex{},
		loader: func() (string, error) {
			return "", errors.New("pac loader has not been configured")
		},
	}
	for _, opt := range opts {
		opt(otto)
	}
	return otto
}

// OttoEngine struct
type OttoEngine struct {
	mutex     *sync.RWMutex
	loader    Loader
	isStarted bool
	vm        *otto.Otto
}

func (o *OttoEngine) Start() error {
	defer func() {
		o.mutex.RLock()
		defer o.mutex.RUnlock()
		if o.isStarted {
			log.Print("started OttoEngine")
		} else {
			log.Print("failed to start OttoEngine")
		}
	}()
	o.mutex.RLock()
	if o.isStarted {
		defer o.mutex.RUnlock()
		return nil
	}
	o.mutex.RUnlock()
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.isStarted {
		return nil
	}
	log.Print("initialising OttoEngine")
	vm := otto.New()

	// ConvertAddr(ipaddr string)
	vm.Set("convert_addr", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			ipaddr string
			err    error
		)
		if ipaddr, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.ConvertAddr(ipaddr)); err == nil {
			value = v
		}
		return
	})

	// DNSDomainIs(host, domain string) bool
	vm.Set("dnsDomainIs", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			host   string
			domain string
			err    error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if domain, err = call.Argument(1).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.DNSDomainIs(host, domain)); err == nil {
			value = v
		}
		return
	})

	// ShExpMatch(str, shexp string) bool
	vm.Set("shExpMatch", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			str   string
			shexp string
			err   error
		)
		if str, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if shexp, err = call.Argument(1).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.ShExpMatch(str, shexp)); err == nil {
			value = v
		}
		return
	})

	// IsInNet(host, netip, netmask string) bool
	vm.Set("isInNet", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			host    string
			netip   string
			netmask string
			err     error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if netip, err = call.Argument(1).ToString(); err != nil {
			return
		}
		if netmask, err = call.Argument(2).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.IsInNet(host, netip, netmask)); err == nil {
			value = v
		}
		return
	})

	// MyIPAddress() string
	vm.Set("myIpAddress", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.NullValue()
		if v, err := vm.ToValue(pacfunc.MyIPAddress()); err == nil {
			value = v
		}
		return
	})

	// DNSResolve(host string) string
	vm.Set("dnsResolve", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			host string
			err  error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.DNSResolve(host)); err == nil {
			value = v
		}
		return
	})

	// IsPlainHostName(host string) bool
	vm.Set("isPlainHostName", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			host string
			err  error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.IsPlainHostName(host)); err == nil {
			value = v
		}
		return
	})

	// LocalHostOrDomainIs(host, hostdom string) bool
	vm.Set("localHostOrDomainIs", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			host    string
			hostdom string
			err     error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if hostdom, err = call.Argument(1).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.LocalHostOrDomainIs(host, hostdom)); err == nil {
			value = v
		}
		return
	})

	// IsResolvable(host string) bool
	vm.Set("isResolvable", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			host string
			err  error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.IsResolvable(host)); err == nil {
			value = v
		}
		return
	})

	// DNSDomainLevels(host string) int
	vm.Set("dnsDomainLevels", func(call otto.FunctionCall) (value otto.Value) {
		value, _ = otto.ToValue(0)
		var (
			host string
			err  error
		)
		if host, err = call.Argument(0).ToString(); err != nil {
			return
		}
		if v, err := vm.ToValue(pacfunc.DNSDomainLevels(host)); err == nil {
			value = v
		}
		return
	})

	// WeekdayRange(wd1, wd2, gmt string) bool
	vm.Set("weekdayRange", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		var (
			wd1, wd2, gmt string
		)
		if call.Argument(0).IsDefined() {
			wd1, _ = call.Argument(0).ToString()
		} else {
			wd1 = ""
		}
		if call.Argument(1).IsDefined() {
			wd2, _ = call.Argument(1).ToString()
		} else {
			wd2 = ""
		}
		if call.Argument(2).IsDefined() {
			gmt, _ = call.Argument(2).ToString()
		} else {
			gmt = ""
		}
		if v, err := vm.ToValue(pacfunc.WeekdayRange(wd1, wd2, gmt)); err == nil {
			value = v
		}
		return
	})

	// DateRange(args []string) bool
	vm.Set("dateRange", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		args := make([]string, len(call.ArgumentList))
		for i := 0; i < len(call.ArgumentList); i++ {
			args[i], _ = call.ArgumentList[i].ToString()
		}
		if v, err := vm.ToValue(pacfunc.DateRange(args)); err == nil {
			value = v
		}
		return
	})

	// TimeRange(args []string) bool
	vm.Set("timeRange", func(call otto.FunctionCall) (value otto.Value) {
		value = otto.FalseValue()
		args := make([]string, len(call.ArgumentList))
		for i := 0; i < len(call.ArgumentList); i++ {
			args[i], _ = call.ArgumentList[i].ToString()
		}
		if v, err := vm.ToValue(pacfunc.TimeRange(args)); err == nil {
			value = v
		}
		return
	})

	{
		pac, pacError := o.loader()
		if pacError != nil {
			return pacError
		}
		log.Print("PAC:\n" + pac + "\n")
		_, pacError = vm.Run(pac)
		if pacError != nil {
			return pacError
		}
	}

	o.vm = vm
	o.isStarted = true

	return nil
}

func (o *OttoEngine) Stop() error {
	defer func() {
		o.mutex.RLock()
		defer o.mutex.RUnlock()
		if !o.isStarted {
			log.Print("stopped OttoEngine")
		} else {
			log.Print("failed to stop OttoEngine")
		}
	}()
	o.mutex.RLock()
	if !o.isStarted {
		defer o.mutex.RUnlock()
		return nil
	}
	o.mutex.RUnlock()
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if !o.isStarted {
		return nil
	}
	log.Print("stopping OttoEngine")
	o.vm = nil
	o.isStarted = false
	return nil
}

func (o *OttoEngine) Reload() error {
	if err := o.Stop(); err != nil {
		return err
	}
	if err := o.Start(); err != nil {
		return err
	}
	return nil
}

func (o *OttoEngine) FindProxyForURL(in *url.URL) (Proxies, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	value, err := o.vm.Call("FindProxyForURL", nil, in.String(), in.Hostname())
	if err != nil {
		return Proxies{}, err
	}

	findProxyString, err := otto.Value.ToString(value)
	if err != nil {
		return Proxies{}, err
	}

	return ParseFindProxyString(findProxyString)
}
