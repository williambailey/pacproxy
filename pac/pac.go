package pac

import (
	"bytes"
	"fmt"
	"net/url"
)

// DirectPAC for simply always returning "DIRECT"
const DirectPAC = "function FindProxyForURL(url, host){ return 'DIRECT'; }"

// EngineManager to deal with the starting, stoping, reloading of any runtime
type EngineManager interface {
	Start() error
	Stop() error
	Reload() error
}

// Proxies is a slice of Proxy that implements Stringer
type Proxies []Proxy

func (p Proxies) String() string {
	var b bytes.Buffer
	for k, v := range p {
		if k > 0 {
			b.Write([]byte("; "))
		}
		b.WriteString(v.String())
	}
	return b.String()
}

// Loader to load the pac as a string
type Loader func() (string, error)

// ProxyFinder for chosing the proxy for a URL
type ProxyFinder interface {
	FindProxyForURL(in *url.URL) (Proxies, error)
}

// ProxySelector for proxy selection
type ProxySelector interface {
	SelectProxy(from Proxies) Proxy
}

/* TODO: Look again at handling multiple proxy entries better...
// ProxyChecker is used when trying to decide which proxy one might use
type ProxyChecker interface {
	IsHealthy(p Proxy) bool
	RecordSuccess(p Proxy)
	RecordFailure(p Proxy)
	ResetRecord(p Proxy)
	ResetAll()
}
*/

// Proxy information struct
type Proxy struct {
	Hostname string
	Port     int
}

func (p Proxy) String() string {
	if p == DirectProxy {
		return "DIRECT"
	}
	return fmt.Sprintf("PROXY %s:%d", p.Hostname, p.Port)
}

// DirectProxy is used to represent a "DIRECT" value
var DirectProxy = Proxy{}
