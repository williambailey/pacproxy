package pacfunc

import (
	"testing"
)

func TestConvertAddr(t *testing.T) {
	assertTrue := func(i string, d uint32) {
		v := ConvertAddr(i)
		if v != d {
			t.Errorf("%q should convert to %d, got %d", i, d, v)
		}
	}
	assertTrue("127.0.0.1", 2130706433)
	assertTrue("10.56.23.193", 171448257)
	assertTrue("0:0:0:0:0:0:7f00:1", 2130706433)
	assertTrue("2000:4A2B::1f3F", 7999)
}

func TestDNSDomainIs(t *testing.T) {
	if !DNSDomainIs("www.netscape.com", ".netscape.com") {
		t.Error("'www.netscape.com' should be a valid host for domain '.netscape.com'")
	}
	if DNSDomainIs("www", ".netscape.com") {
		t.Error("'www' should not be a valid host for domain '.netscape.com'")
	}
	if DNSDomainIs("www.mcom.com", ".netscape.com") {
		t.Error("'www.mcom.com' should not be a valid host for domain '.netscape.com'")
	}
}

func TestShExpMatch(t *testing.T) {
	if !ShExpMatch("http://home.netscape.com/people/ari/index.html", "*/ari/*") {
		t.Error("'http://home.netscape.com/people/ari/index.html' should match '*/ari/*'")
	}
	if ShExpMatch("http://home.netscape.com/people/montulli/index.html", "*/ari/*") {
		t.Error("'http://home.netscape.com/people/montulli/index.html' should not match '*/ari/*'")
	}
}

func TestIsInNet(t *testing.T) {
	assertTrue := func(h, i, m string) {
		if !IsInNet(h, i, m) {
			t.Errorf("%q should fall within the network %q with the mask %q", h, i, m)
		}
	}
	assertFalse := func(h, i, m string) {
		if IsInNet(h, i, m) {
			t.Errorf("%q should not fall within the network %q with the mask %q", h, i, m)
		}
	}
	assertFalse("", "172.16.0.0", "255.240.0.0")
	assertFalse("unresolvable.example.com", "172.16.0.0", "255.240.0.0")
	assertTrue("172.16.0.1", "172.16.0.0", "255.240.0.0")
	assertFalse("172.1.0.1", "172.16.0.0", "255.240.0.0")
	assertTrue("localhost", "127.0.0.0", "255.0.0.0")
	assertTrue("localhost", "127.1.2.3", "255.0.0.0")
	assertFalse("192.168.1.23", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.24", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.25", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.26", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.27", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.28", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.29", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.30", "192.168.1.24", "255.255.255.248")
	assertTrue("192.168.1.31", "192.168.1.24", "255.255.255.248")
	assertFalse("192.168.1.32", "192.168.1.24", "255.255.255.248")
}

func TestMyIPAddress(t *testing.T) {
	// Hmm. Testing this would simply be reproducing the function
}

func TestDNSResolve(t *testing.T) {
	ip := DNSResolve("localhost")
	if ip != "127.0.0.1" {
		t.Errorf("Expecting localhost to resolve to 127.0.0.1, got %q", ip)
	}
	ip = DNSResolve("unresolvable.example.com")
	if ip != "" {
		t.Errorf("Expecting unresolvable.example.com not resolve, got %q", ip)
	}
}

func TestIsPlainHostName(t *testing.T) {
	if !IsPlainHostName("internet") {
		t.Error("Expecting \"internet\" to be classes as a plan hostname")
	}
	if IsPlainHostName("inter.net") {
		t.Error("Expecting \"inter.net\" to not be classes as a plan hostname")
	}
}

func TestLocalHostOrDomainIs(t *testing.T) {
	assertTrue := func(h, d string) {
		if !LocalHostOrDomainIs(h, d) {
			t.Errorf("expecting %q to be true for %q", h, d)
		}
	}
	assertFalse := func(h, d string) {
		if LocalHostOrDomainIs(h, d) {
			t.Errorf("expecting %q to be false for %q", h, d)
		}
	}
	assertTrue("www.example.com", "www.example.com")
	assertTrue("www.example", "www.example.com")
	assertTrue("www", "www.example.com")
	assertFalse("www", "example.com")
	assertFalse("ftp", "www.example.com")
}

func TestIsResolvable(t *testing.T) {
	assertTrue := func(h string) {
		if !IsResolvable(h) {
			t.Errorf("expecting %q to be true", h)
		}
	}
	assertFalse := func(h string) {
		if IsResolvable(h) {
			t.Errorf("expecting %q to be false", h)
		}
	}
	assertFalse("")
	assertTrue("localhost")
	assertFalse("unresolvable.example.com")
}

func TestDNSDomainLevels(t *testing.T) {
	assertLevels := func(h string, i int) {
		j := DNSDomainLevels(h)
		if i != j {
			t.Errorf("expecting %q to have %d levels, got %d", h, i, j)
		}
	}
	assertLevels("localhost", 0)
	assertLevels("local.host", 1)
	assertLevels("www.example.org", 2)
	assertLevels("a.b.c.d.example.org", 5)
}
