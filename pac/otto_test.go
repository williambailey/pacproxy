package pac

import (
	"net/url"
	"testing"
)

func assertOtto(t *testing.T, pac string, u string, p []Proxy, e string) {
	o := NewOttoEngine(
		OttoStringLoader(pac),
	)
	if err := o.Start(); err != nil {
		t.Errorf("failed to start otto: %q", err)
		return
	}
	assertOttoFind(t, o, u, p, e)
	if err := o.Stop(); err != nil {
		t.Errorf("failed to stop otto: %q", err)
		return
	}
}

func assertOttoFind(t *testing.T, o *OttoEngine, u string, p []Proxy, e string) {
	url, err := url.Parse(u)
	if err != nil {
		t.Errorf("failed to parse url: %q", err)
		return
	}
	proxies, err := o.FindProxyForURL(url)
	if e == "" {
		if err != nil {
			t.Errorf("unexpected error: %q", err)
		}
	} else {
		if err == nil {
			t.Errorf("expecting error %q, got nil", e)
		} else if err.Error() != e {
			t.Errorf("expecting error %q, got %q", e, err)
		}
	}
	if len(p) != len(proxies) {
		t.Errorf("expecting %d proxies, got %d", len(p), len(proxies))
	}
	maxI := len(p)
	if len(proxies) > maxI {
		maxI = len(proxies)
	}
	for i := 0; i < maxI; i++ {
		a := "undefined"
		b := "undefined"
		if len(p) > i {
			a = p[i].String()
		}
		if len(proxies) > i {
			b = proxies[i].String()
		}
		if a != b {
			t.Errorf("expecting proxy item at offset %d to be %q, got %q", i, a, b)
		}
	}
}

func TestOttoWithoutPac(t *testing.T) {
	o := NewOttoEngine()
	err := o.Start()
	if err == nil {
		t.Errorf("expecting an error on start")
		return
	}
	e := "pac loader has not been configured"
	if err.Error() != e {
		t.Errorf("expecting error %q, got %q", e, err)
	}
}

func TestOttoFindProxyForURL(t *testing.T) {
	assertOtto(
		t,
		"function FindProxyForURL(url, host){ return 1234; }",
		"http://www.example.com/page.html",
		[]Proxy{},
		"unsupported PAC command \"1234\"",
	)
}

func TestOttoWithUndefinedFindProxyForURLFunction(t *testing.T) {
	assertOtto(
		t,
		"1 + 1",
		"http://www.example.com/page.html",
		[]Proxy{},
		"ReferenceError: 'FindProxyForURL' is not defined",
	)
}

func TestOttoWithNonFindProxyForURLFunction(t *testing.T) {
	assertOtto(
		t,
		"FindProxyForURL = 1234",
		"http://www.example.com/page.html",
		[]Proxy{},
		"TypeError: 'FindProxyForURL' is not a function",
	)
}

func TestOttoWithFindProxyForURLFunctionThatReturnsInvalidValue(t *testing.T) {
	assertOtto(
		t,
		"function FindProxyForURL(url, host){ return 1234; }",
		"http://www.example.com/page.html",
		[]Proxy{},
		"unsupported PAC command \"1234\"",
	)
}

func TestOttoWithDirectPAC(t *testing.T) {
	assertOtto(
		t,
		DirectPAC,
		"http://www.example.com/page.html",
		[]Proxy{DirectProxy},
		"",
	)
}

func TestOttoWithFindProxyForURLFunctionThatReturnsMultipleValues(t *testing.T) {
	assertOtto(
		t,
		"function FindProxyForURL(url, host){ return 'PROXY proxy.example.com:8080; DIRECT'; }",
		"http://www.example.com/page.html",
		[]Proxy{Proxy{"proxy.example.com", 8080}, DirectProxy},
		"",
	)
}
