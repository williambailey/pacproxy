package main

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacLoad(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(pacDefaultJavascript))
}

func TestPacLoadStringReturnErrorWhenGivenInvalidJavascript(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.Error(t, p.Load(pacDefaultJavascript+"={;}"))
}

func TestPacCallFindProxyForURLRevertsToDefaultWhenNothingHasBeenLoaded(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	s, e := p.CallFindProxyForURL("http://example.com/foo", "example.com")
	require.NoError(t, e)
	require.Equal(t, "DIRECT", s)
}

func TestPacCallFindProxyForURLWithLoadedPacAndThenUnloadAndTryAgain(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(`
function FindProxyForURL(url, host)
{
	return "PROXY example.com";
}`))
	s, e := p.CallFindProxyForURL("http://example.com/foo", "example.com")
	require.NoError(t, e)
	require.Equal(t, "PROXY example.com", s)
	p.Unload()
	s, e = p.CallFindProxyForURL("http://example.com/foo", "example.com")
	require.NoError(t, e)
	require.Equal(t, "DIRECT", s)
}

func TestPacProxyReturnsNilURLAndNoErrorWhenPassedNilURL(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	u, e := p.Proxy(nil)
	require.NoError(t, e)
	require.Nil(t, u)
}

func TestPacProxyReturnsNilURLAndNoErrorForDirectResult(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	u, e := p.Proxy(&url.URL{})
	require.NoError(t, e)
	require.Nil(t, u)
}

func TestPacProxyReturnsNilURLAndErrorForProxyResultWhenConnectionFails(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(`
function FindProxyForURL(url, host)
{
	return "PROXY 127.0.0.1:9";
}`))
	in, _ := url.Parse("http://example.com:82/foo")
	out, e := p.Proxy(in)
	require.Error(t, e)
	require.Nil(t, out)
	require.Contains(t, e.Error(), "Unable to process FindProxyForURL(\"http://example.com:82/foo\", \"example.com\") result \"PROXY 127.0.0.1:9\".\nConnection to \"127.0.0.1:9\" is currently blacklisted for 4m59.")
	require.Contains(t, e.Error(), "s: dial tcp 127.0.0.1:9: connection refused.")
}

func TestPacProxyReturnsNilURLAndNilErrorForProxyResultWhenConnectionFailsButWeHaveDirectFallback(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(`
function FindProxyForURL(url, host)
{
	return "PROXY 127.0.0.1:9; DIRECT";
}`))
	in, _ := url.Parse("http://example.com:82/foo")
	out, e := p.Proxy(in)
	require.NoError(t, e)
	require.Nil(t, out)
}

func TestPacPacConfiguration(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.Equal(t, pacDefaultJavascript, string(p.PacConfiguration()))
	e = p.LoadFile("./resource/test/example.pac")
	require.NoError(t, e)
	f, _ := ioutil.ReadFile("./resource/test/example.pac")
	require.Equal(t, string(f), string(p.PacConfiguration()))
}

func TestPacPacFilename(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.Equal(t, "", p.PacFilename())
	e = p.LoadFile("./resource/test/example.pac")
	require.NoError(t, e)
	f, _ := filepath.Abs("./resource/test/example.pac")
	require.Equal(t, f, p.PacFilename())
}
