package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

type customResolver struct {
	resolver *net.Resolver
	search   []string
}

// newCustomResolver returns nil when servers is empty so callers can
// fall back to the system resolver.
func newCustomResolver(servers, search []string) *customResolver {
	if len(servers) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(servers))
	for _, s := range servers {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, _, err := net.SplitHostPort(s); err != nil {
			s = net.JoinHostPort(s, "53")
		}
		normalized = append(normalized, s)
	}
	if len(normalized) == 0 {
		return nil
	}

	cleanSearch := make([]string, 0, len(search))
	for _, d := range search {
		d = strings.TrimSpace(strings.Trim(d, "."))
		if d != "" {
			cleanSearch = append(cleanSearch, d)
		}
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	var idx int
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			server := normalized[idx%len(normalized)]
			idx++
			return dialer.DialContext(ctx, network, server)
		},
	}

	return &customResolver{resolver: r, search: cleanSearch}
}

// resolveHost applies search domains to single-label names.
// IPs pass through unchanged.
func (c *customResolver) resolveHost(ctx context.Context, hostPort string) (string, error) {
	host, port, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", fmt.Errorf("split %q: %w", hostPort, err)
	}
	if ip := net.ParseIP(host); ip != nil {
		return hostPort, nil
	}

	candidates := []string{host}
	if !strings.Contains(host, ".") {
		for _, d := range c.search {
			candidates = append(candidates, host+"."+d)
		}
	}

	var lastErr error
	for _, name := range candidates {
		addrs, err := c.resolver.LookupHost(ctx, name)
		if err == nil && len(addrs) > 0 {
			return net.JoinHostPort(addrs[0], port), nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no addresses returned for %q", host)
	}
	return "", lastErr
}

func (c *customResolver) dialContext(d *net.Dialer) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		resolved, err := c.resolveHost(ctx, address)
		if err != nil {
			return nil, err
		}
		return d.DialContext(ctx, network, resolved)
	}
}
