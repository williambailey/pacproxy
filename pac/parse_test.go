package pac

import (
	"errors"
	"fmt"
	"testing"
)

var parsetests = []struct {
	in      string
	proxies []Proxy
	err     error
}{
	{"DIRECT", []Proxy{DirectProxy}, nil},
	{"PROXY proxy.example.com:8080", []Proxy{Proxy{"proxy.example.com", 8080}}, nil},
	{"PROXY proxy.example.com:8080;", []Proxy{Proxy{"proxy.example.com", 8080}}, nil},
	{"PROXY proxy.example.com:8080;  ; ;;", []Proxy{Proxy{"proxy.example.com", 8080}}, nil},
	{"PROXY proxy.example.com:8080; DIRECT", []Proxy{Proxy{"proxy.example.com", 8080}, DirectProxy}, nil},
	{"PROXY proxy.example.com:8080; DIRECT; PROXY proxy.example.org:8888", []Proxy{Proxy{"proxy.example.com", 8080}, DirectProxy, Proxy{"proxy.example.org", 8888}}, nil},
	{"FOO", []Proxy{}, errors.New("unsupported PAC command \"FOO\"")},
	{"PROXY", []Proxy{}, errors.New("unable to parse proxy details from \"PROXY\"")},
	{"PROXY http://foo.bar:8080", []Proxy{}, errors.New("unable to parse hostname and port from \"http://foo.bar:8080\"")},
	{"PROXY proxy.example.com", []Proxy{}, errors.New("unable to parse hostname and port from \"proxy.example.com\"")},
}

func TestParseFindProxyString(t *testing.T) {
	for _, pt := range parsetests {
		proxies, err := ParseFindProxyString(pt.in)
		if fmt.Sprintf("%q", err) != fmt.Sprintf("%q", pt.err) {
			t.Errorf("%q error expected %q, got %q", pt.in, pt.err, err)
		}
		if len(proxies) != len(pt.proxies) {
			t.Errorf("%q expected %d proxies, got %d", pt.in, len(pt.proxies), len(proxies))
		} else {
			for k, v := range pt.proxies {
				if proxies[k] != v {
					t.Errorf("%q expected proxy %d to be %q, got %q", pt.in, k, v, proxies[k])
				}
			}
		}
	}
}
