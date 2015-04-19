package main

import (
	"net/url"
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

func TestPacFindProxyForURLRevertsToDefaultWhenNothingHasBeenLoaded(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	s, e := p.CallFindProxyForURL("http://example.com/foo", "example.com")
	require.NoError(t, e)
	require.Equal(t, "DIRECT", s)
}

func TestPacFindProxyWithLoadedPacAndThenUnloadAndTryAgain(t *testing.T) {
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

func TestPacFindProxyReturnsNilURLAndNoErrorWhenPassedNilURL(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	u, e := p.FindProxy(nil)
	require.NoError(t, e)
	require.Nil(t, u)
}

func TestPacFindProxyReturnsNilURLAndNoErrorForDirectResult(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	u, e := p.FindProxy(&url.URL{})
	require.NoError(t, e)
	require.Nil(t, u)
}

func TestPacFindProxyReturnsURLAndNoErrorForProxyResult(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(`
function FindProxyForURL(url, host)
{
	return "PROXY example.com";
}`))
	in, _ := url.Parse("http://example.com")
	out, e := p.FindProxy(in)
	require.NoError(t, e)
	require.NotNil(t, out)
	require.Equal(t, "myproxy.com:8080", out.String())
}
