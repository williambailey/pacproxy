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

//package gopac
package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGopacUtilsIsPlainHostName(t *testing.T) {
	assert.True(t, gopacIsPlainHostName("www"), "'www' should be a valid plain host")
	assert.False(t, gopacIsPlainHostName("www.netscape.com"), "'www.netscape.com' should not be a valid plain host")
}

func TestGopacUtilsDnsDomainIs(t *testing.T) {
	assert.True(t, gopacDnsDomainIs("www.netscape.com", ".netscape.com"), "'www.netscape.com' should be a valid host for domain '.netscape.com'")
	assert.False(t, gopacDnsDomainIs("www", ".netscape.com"), "'www' should not be a valid host for domain '.netscape.com'")
	assert.False(t, gopacDnsDomainIs("www.mcom.com", ".netscape.com"), "'www.mcom.com' should not be a valid host for domain '.netscape.com'")
}

func TestGopacUtilsLocalHostOrDomainIs(t *testing.T) {
	assert.True(t, gopacLocalHostOrDomainIs("www.netscape.com", "www.netscape.com"), "'www.netscape.com' should be valid as it equals the domain 'www.netscape.com'")
	assert.True(t, gopacLocalHostOrDomainIs("www", "www.netscape.com"), "'www' should be valid as it contains no domain part")
	assert.False(t, gopacLocalHostOrDomainIs("www.mcom.com", "wwww.netscape.com"), "'www.mcom.com' should not be as it contains a domain part")
	assert.False(t, gopacLocalHostOrDomainIs("home.netscape.com", "wwww.netscape.com"), "'home.netscape.com' should not be as it contains a domain part")
}

func TestGopacUtilsIsResolvable(t *testing.T) {
	assert.True(t, gopacIsInNet("localhost", "127.0.0.1", "255.255.255.255"), "'localhost' should equal 127.0.0.1 with the mask 255.255.255.255")
	assert.True(t, gopacIsInNet("localhost", "127.0.0.0", "255.0.0.0"), "'localhost' should equal 127.0.0.1 with the mask 255.0.0.0")
	assert.False(t, gopacIsInNet("localhost", "127.0.0.0", "255.0.0.255"), "'localhost' should not equal 127.0.0.1 with the mask 255.0.0.255")
}

func TestGopacUtilsDnsResolve(t *testing.T) {
	assert.Equal(t, gopacDnsResolve("localhost"), "127.0.0.1", "'localhost' should equal 127.0.0.1")
}

func TestGopacUtilsDnsDomainLevels(t *testing.T) {
	assert.Equal(t, gopacDnsDomainLevels("www"), 0, "'www' should contain 0 domain levels")
	assert.Equal(t, gopacDnsDomainLevels("www.netscape.com"), 2, "'www.netscape.com' should contain 2 domain levels")
}

func TestGopacUtilsShExpMatch(t *testing.T) {
	assert.True(t, gopacShExpMatch("http://home.netscape.com/people/ari/index.html", "*/ari/*"), "'http://home.netscape.com/people/ari/index.html' should match '*/ari/*'")
	assert.False(t, gopacShExpMatch("http://home.netscape.com/people/montulli/index.html", "*/ari/*"), "'http://home.netscape.com/people/montulli/index.html' should not match '*/ari/*'")
}
