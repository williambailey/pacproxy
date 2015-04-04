// Copyright 2014 Jack Wakefield
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gopac

import (
	"testing"

	"github.com/robertkrimen/otto"
	"github.com/stretchr/testify/assert"
)

func callBooleanFunction(rt *runtime, method string, params ...interface{}) (bool, error) {
	value, err := rt.vm.Call(method, nil, params...)

	if err != nil {
		return false, err
	}

	return otto.Value.ToBoolean(value)
}

func callStringFunction(rt *runtime, method string, params ...interface{}) (string, error) {
	value, err := rt.vm.Call(method, nil, params...)

	if err != nil {
		return "", err
	}

	return otto.Value.ToString(value)
}

func callNumberFunction(rt *runtime, method string, params ...interface{}) (int64, error) {
	value, err := rt.vm.Call(method, nil, params...)

	if err != nil {
		return 0, err
	}

	return otto.Value.ToInteger(value)
}

func TestRuntimeInit(t *testing.T) {
	rt, err := newRuntime()
	assert.NotNil(t, rt, "runtime should not be nil")
	assert.Nil(t, err, "err should be nil")
}

func TestRuntimeIsPlainHostName(t *testing.T) {
	rt, _ := newRuntime()

	www, err := callBooleanFunction(rt, "isPlainHostName", "www")
	assert.Nil(t, err, "should not error")
	assert.True(t, www, "'www' should be a valid plain host")

	netscape, err := callBooleanFunction(rt, "isPlainHostName", "www.netscape.com")
	assert.Nil(t, err, "should not error")
	assert.False(t, netscape, "'www.netscape.com' should not be a valid plain host")
}

func TestRuntimeDnsDomainIs(t *testing.T) {
	rt, _ := newRuntime()

	netscape, err := callBooleanFunction(rt, "dnsDomainIs", "www.netscape.com", ".netscape.com")
	assert.Nil(t, err, "should not error")
	assert.True(t, netscape, "'www.netscape.com' should be a valid host for domain '.netscape.com'")

	www, err := callBooleanFunction(rt, "dnsDomainIs", "www", ".netscape.com")
	assert.Nil(t, err, "should not error")
	assert.False(t, www, "'www' should not be a valid host for domain '.netscape.com'")

	mcom, err := callBooleanFunction(rt, "dnsDomainIs", "w.mcom.com", ".netscape.com")
	assert.Nil(t, err, "should not error")
	assert.False(t, mcom, "'www.mcom.com' should not be a valid host for domain '.netscape.com'")
}

func TestRuntimeLocalHostOrDomainIs(t *testing.T) {
	rt, _ := newRuntime()

	netscape, err := callBooleanFunction(rt, "localHostOrDomainIs", "www.netscape.com", "www.netscape.com")
	assert.Nil(t, err, "should not error")
	assert.True(t, netscape, "'www.netscape.com' should be valid as it equals the domain 'www.netscape.com'")

	www, err := callBooleanFunction(rt, "localHostOrDomainIs", "www", "www.netscape.com")
	assert.Nil(t, err, "should not error")
	assert.True(t, www, "'www' should be valid as it contains no domain part")

	mcom, err := callBooleanFunction(rt, "localHostOrDomainIs", "www.mcom.com", "wwww.netscape.com")
	assert.Nil(t, err, "should not error")
	assert.False(t, mcom, "'www.mcom.com' should not be as it contains a domain part")

	home, err := callBooleanFunction(rt, "localHostOrDomainIs", "home.netscape.com", "wwww.netscape.com")
	assert.Nil(t, err, "should not error")
	assert.False(t, home, "'home.netscape.com' should not be as it contains a domain part")
}

func TestRuntimeIsResolvable(t *testing.T) {
	rt, _ := newRuntime()

	localhost1, err := callBooleanFunction(rt, "isInNet", "localhost", "127.0.0.1", "255.255.255.255")
	assert.Nil(t, err, "should not error")
	assert.True(t, localhost1, "'localhost' should equal 127.0.0.1 with the mask 255.255.255.255")

	localhost2, err := callBooleanFunction(rt, "isInNet", "localhost", "127.0.0.0", "255.0.0.0")
	assert.Nil(t, err, "should not error")
	assert.True(t, localhost2, "'localhost' should equal 127.0.0.1 with the mask 255.0.0.0")

	localhost3, err := callBooleanFunction(rt, "isInNet", "localhost", "127.0.0.0", "255.0.0.255")
	assert.Nil(t, err, "should not error")
	assert.False(t, localhost3, "'localhost' should not equal 127.0.0.1 with the mask 255.0.0.255")
}

func TestRuntimeDnsResolve(t *testing.T) {
	rt, _ := newRuntime()

	localhost, err := callStringFunction(rt, "dnsResolve", "localhost")
	assert.Nil(t, err, "should not error")
	assert.Equal(t, localhost, "127.0.0.1", "'localhost' should equal 127.0.0.1")
}

func TestRuntimeDnsDomainLevels(t *testing.T) {
	rt, _ := newRuntime()

	www, err := callNumberFunction(rt, "dnsDomainLevels", "www")
	assert.Nil(t, err, "should not error")
	assert.Equal(t, www, 0, "'www' should contain 0 domain levels")

	netscape, err := callNumberFunction(rt, "dnsDomainLevels", "www.netscape.com")
	assert.Nil(t, err, "should not error")
	assert.Equal(t, netscape, 2, "'www.netscape.com' should contain 2 domain levels")
}

func TestRuntimeShExpMatch(t *testing.T) {
	rt, _ := newRuntime()

	ari, err := callBooleanFunction(rt, "shExpMatch", "http://home.netscape.com/people/ari/index.html", "*/ari/*")
	assert.Nil(t, err, "should not error")
	assert.True(t, ari, "'http://home.netscape.com/people/ari/index.html' should match '*/ari/*'")

	montulli, err := callBooleanFunction(rt, "shExpMatch", "http://home.netscape.com/people/montulli/index.html", "*/ari/*")
	assert.Nil(t, err, "should not error")
	assert.False(t, montulli, "'http://home.netscape.com/people/montulli/index.html' should not match '*/ari/*'")
}
