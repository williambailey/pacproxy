package pacfunc

import (
	"encoding/binary"
	"net"
	"os"
	"regexp"
	"strings"
)

// ConvertAddr converts an IPv4 dotted decimal IP address or an IPv6 IP address to an integer
func ConvertAddr(ipaddr string) uint32 {
	ip := net.ParseIP(ipaddr)
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

// DNSDomainIs evaluates hostnames and returns true if hostnames match.
//
// Used mainly to match and exception individual hostnames.
func DNSDomainIs(host, domain string) bool {
	if len(host) < len(domain) {
		return false
	}
	return strings.HasSuffix(host, domain)
}

// ShExpMatch will attempt to match hostname or URL to a specified shell expression, and returns true if matched.
func ShExpMatch(str, shexp string) bool {
	shexp = strings.Replace(shexp, ".", "\\.", -1)
	shexp = strings.Replace(shexp, "?", ".?", -1)
	shexp = strings.Replace(shexp, "*", ".*", -1)
	matched, err := regexp.MatchString("^"+shexp+"$", str)
	return err == nil && matched
}

// IsInNet evaluates the IP address of a hostname, and if within a specified
// subnet returns true. If a hostname is passed the function will resolve the
// hostname to an IP address.
func IsInNet(host, netip, netmask string) bool {
	if len(host) == 0 {
		return false
	}
	address, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return false
	}
	net := net.IPNet{
		IP:   net.ParseIP(netip),
		Mask: net.IPMask(net.ParseIP(netmask)),
	}
	return net.Contains(address.IP)
}

// MyIPAddress returns the IP address of the host machine.
func MyIPAddress() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "127.0.0.1"
	}
	return DNSResolve(hostname)
}

// DNSResolve returns the IP address of the host.
func DNSResolve(host string) string {
	address, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return ""
	}
	return address.String()
}

// IsPlainHostName will return true if the hostname contains no dots, e.g. http://intranet
//
// Useful when applying exceptions for internal websites, e.g. may not require
// resolution of a hostname to IP address to determine if local.
func IsPlainHostName(host string) bool {
	return strings.Index(host, ".") == -1
}

// LocalHostOrDomainIs evaluates hostname and only returns true if exact
// hostname match is found.
func LocalHostOrDomainIs(host, hostdom string) bool {
	if host == hostdom {
		return true
	}
	return strings.LastIndex(hostdom, host+".") == 0
}

// IsResolvable attempts to resolve a hostname to an IP address and returns
// true if successful.
func IsResolvable(host string) bool {
	if len(host) == 0 {
		return false
	}
	if _, err := net.ResolveIPAddr("ip", host); err != nil {
		return false
	}
	return true
}

// DNSDomainLevels returns the number of DNS domain levels (number of dots)
// in the hostname. Can be used to exception internal websites which use short
// DNS names, e.g. http://intranet
func DNSDomainLevels(host string) int {
	return strings.Count(host, ".")
}
