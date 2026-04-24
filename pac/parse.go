package pac

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	pacStatementSplit *regexp.Regexp
	pacItemSplit      *regexp.Regexp
)

func init() {
	pacStatementSplit = regexp.MustCompile(`\s*;\s*`)
	pacItemSplit = regexp.MustCompile(`\s+`)
}

// ParseFindProxyString into a Proxies
func ParseFindProxyString(s string) (Proxies, error) {
	// "PROXY proxy.example.com:8080; DIRECT"
	var (
		proxies     Proxies
		proxyURL    *url.URL
		proxyURLErr error
		hostname    string
		portStr     string
		portInt     int
		portErr     error
	)
	for _, statement := range pacStatementSplit.Split(s, 50) {
		if statement == "" {
			continue
		}
		part := pacItemSplit.Split(statement, 2)
		switch strings.ToUpper(part[0]) {
		case "DIRECT":
			proxies = append(proxies, DirectProxy)
		case "PROXY", "HTTPS":
			if len(part) != 2 {
				return Proxies{}, fmt.Errorf("unable to parse proxy details from %q", statement)
			}
			proxyURL, proxyURLErr = url.Parse("http://" + part[1])
			if proxyURLErr != nil {
				return Proxies{}, proxyURLErr
			}
			hostname = proxyURL.Hostname()
			portStr = proxyURL.Port()
			if hostname == "" || portStr == "" {
				return Proxies{}, fmt.Errorf("unable to parse hostname and port from %q", part[1])
			}
			portInt, portErr = strconv.Atoi(proxyURL.Port())
			if portErr != nil {
				return Proxies{}, portErr
			}
			proxies = append(proxies, Proxy{
				Hostname: proxyURL.Hostname(),
				Port:     portInt,
				TLS:      strings.ToUpper(part[0]) == "HTTPS",
			})
		default:
			return Proxies{}, fmt.Errorf("unsupported PAC command %q", part[0])
		}
	}
	return proxies, nil
}
