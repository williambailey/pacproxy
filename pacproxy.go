package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/williambailey/pacproxy/pac"
)

// Name of the app
const Name = "pacproxy"

// Version of the app
const Version = "2.0.7"

// About the app
const About = "A no-frills local HTTP proxy server powered by a proxy auto-config (PAC) file"

// Repo where the app is located
const Repo = "https://github.com/williambailey/pacproxy"

var (
	fPac              string
	fListen           string
	fVerbose          bool
	fResolveURL       string
	fUpstreamUser     string
	fUpstreamPassFile string
	fUpstreamAuthHost string
	fDNS              string
	fDNSSearch        string
)

func init() {
	flag.StringVar(&fPac, "c", "", "PAC file name, url or javascript to use (required)")
	flag.StringVar(&fListen, "l", "127.0.0.1:8080", "Interface and port to listen on")
	flag.BoolVar(&fVerbose, "v", false, "send verbose output to STDERR")
	flag.StringVar(&fResolveURL, "r", "", "Resolve the proxies for the provided url to STDOUT and exit")
	flag.StringVar(&fUpstreamUser, "upstream-user", "", "Username for upstream proxy Basic authentication")
	flag.StringVar(&fUpstreamPassFile, "upstream-password-file", "", "File containing the upstream proxy password")
	flag.StringVar(&fUpstreamAuthHost, "upstream-auth-hosts", "", "Comma-separated list of upstream hosts (with optional :port) that should receive Proxy-Authorization. Empty means all.")
	flag.StringVar(&fDNS, "dns", "", "Comma-separated list of DNS server IPs (with optional :port) to use instead of the system resolver. Empty = use system DNS.")
	flag.StringVar(&fDNSSearch, "dns-search", "", "Comma-separated list of DNS search domains appended to unqualified hostnames. Only used when -dns is set.")
}

func main() {
	flag.Usage = func() {
		// fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "%s v%s\n\n%s\n%s\n\nUsage:\n", Name, Version, About, Repo)
		flag.PrintDefaults()
	}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	required := []string{"c"}
	for _, req := range required {
		if !seen[req] {
			exitWithUsage(fmt.Sprintf("Missing required flag -%s", req))
		}
	}
	if strings.TrimSpace(fPac) == "" {
		exitWithUsage("Unexpected empty value for -c")
	}

	if fVerbose {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.Discard)
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

	if fResolveURL != "" {
		do_resolve(otto)
		return
	}
	listen(otto)
}

func exitWithUsage(message string) {
	os.Stderr.WriteString(message)
	os.Stderr.WriteString("\n")
	flag.Usage()
	os.Exit(2) // the same exit code flag.Parse uses
}

func do_resolve(otto *pac.OttoEngine) {
	u, err := url.Parse(fResolveURL)
	if err != nil {
		log.Printf("Unable to parse resolve URL %q: %s", fResolveURL, err)
		os.Exit(1)
	}
	proxies, err := otto.FindProxyForURL(u)
	if err != nil {
		log.Printf("Error while trying to find proxy for URL %q: %s", fResolveURL, err)
		os.Exit(1)
	}
	for i := 0; i < len(proxies); i++ {
		fmt.Println(proxies[i])
	}
}

func listen(otto *pac.OttoEngine) {
	var upstreamCreds *upstreamAuth
	if fUpstreamUser != "" {
		if fUpstreamPassFile == "" {
			exitWithUsage("-upstream-user requires -upstream-password-file")
		}
		raw, err := os.ReadFile(fUpstreamPassFile)
		if err != nil {
			log.Panicf("unable to read upstream password file %q: %s", fUpstreamPassFile, err)
		}
		pass := strings.TrimRight(string(raw), "\r\n\t ")
		if pass == "" {
			log.Panicf("upstream password file %q is empty", fUpstreamPassFile)
		}
		var hosts []string
		if fUpstreamAuthHost != "" {
			for _, h := range strings.Split(fUpstreamAuthHost, ",") {
				h = strings.TrimSpace(strings.ToLower(h))
				if h != "" {
					hosts = append(hosts, h)
				}
			}
		}
		upstreamCreds = &upstreamAuth{user: fUpstreamUser, pass: pass, hosts: hosts}
		if len(hosts) > 0 {
			log.Printf("Upstream Basic auth enabled for user %q, restricted to hosts %v", fUpstreamUser, hosts)
		} else {
			log.Printf("Upstream Basic auth configured for user %q but no -upstream-auth-hosts set: credentials will NOT be sent anywhere", fUpstreamUser)
		}
	} else if fUpstreamPassFile != "" {
		exitWithUsage("-upstream-password-file requires -upstream-user")
	} else if fUpstreamAuthHost != "" {
		exitWithUsage("-upstream-auth-hosts requires -upstream-user")
	}

	var dnsServers, dnsSearch []string
	if fDNS != "" {
		for _, s := range strings.Split(fDNS, ",") {
			if s = strings.TrimSpace(s); s != "" {
				dnsServers = append(dnsServers, s)
			}
		}
	}
	if fDNSSearch != "" {
		for _, d := range strings.Split(fDNSSearch, ",") {
			if d = strings.TrimSpace(d); d != "" {
				dnsSearch = append(dnsSearch, d)
			}
		}
	}
	if len(dnsServers) == 0 && len(dnsSearch) > 0 {
		exitWithUsage("-dns-search requires -dns")
	}
	resolver := newCustomResolver(dnsServers, dnsSearch)
	if resolver != nil {
		log.Printf("Custom DNS resolver enabled: servers=%v search=%v", dnsServers, dnsSearch)
	}

	srv := &http.Server{
		Addr:              fListen,
		ReadHeaderTimeout: 2 * time.Second,
		IdleTimeout:       60 * time.Second,
		Handler: newProxyHTTPHandler(
			otto,
			&pac.FirstItemSelector{},
			newNonProxyHTTPHandler(),
			upstreamCreds,
			resolver,
		),
	}
	log.Printf("Listening on %q", fListen)
	if err := srv.ListenAndServe(); err != nil {
		log.Panic(err)
	}
}
