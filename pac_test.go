package main

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPacLoad(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(MustAsset("default.pac")))
}

func TestPacLoadStringReturnErrorWhenGivenInvalidJavascript(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.Error(t, p.Load(string(MustAsset("default.pac"))+"={;}"))
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
	require.Contains(t, e.Error(), "Unable to process FindProxyForURL(\"http://example.com:82/foo\", \"example.com\") result \"PROXY 127.0.0.1:9\".\nConnection to \"127.0.0.1:9\" is currently blacklisted for 4m59s:")
	require.Contains(t, e.Error(), "127.0.0.1:9")
	require.Contains(t, e.Error(), "connection refused.")
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
	require.Equal(t, string(MustAsset("default.pac")), string(p.PacConfiguration()))
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

func TestPacCallFindProxyForURLFromGoRoutineDoesNotHaveRaceIssues(t *testing.T) {
	p, e := NewPac()
	require.NoError(t, e)
	require.NoError(t, p.Load(`
function FindProxyForURL(url, host)
{
	return "PROXY " + host;
}`))
	var wg sync.WaitGroup
	max := 1000
	c := make(chan string, max)
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func(i int) {
			s, e := p.CallFindProxyForURL(fmt.Sprintf("http://%d", i), fmt.Sprintf("%d", i))
			c <- s
			wg.Done()
			require.NoError(t, e)
			require.Equal(t, fmt.Sprintf("PROXY %d", i), s)
		}(i)
	}
	inSequance := true
	for i := 0; i < max; i++ {
		s := <-c
		if fmt.Sprintf("PROXY %d", i) != s {
			inSequance = false
		}
	}
	require.Equal(t, inSequance, false)
	wg.Wait()
}
