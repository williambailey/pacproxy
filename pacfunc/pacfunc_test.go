package pacfunc

import (
	"testing"
	"time"
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

var (
	ny, _        = time.LoadLocation("America/New_York")
	sundayUTC    = time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC)
	mondayUTC    = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	tuesdayUTC   = time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)
	wednesdayUTC = time.Date(2018, 1, 3, 0, 0, 0, 0, time.UTC)
	thursdayUTC  = time.Date(2018, 1, 4, 0, 0, 0, 0, time.UTC)
	fridayUTC    = time.Date(2018, 1, 5, 0, 0, 0, 0, time.UTC)
	saturdayUTC  = time.Date(2018, 1, 6, 0, 0, 0, 0, time.UTC)
)

var weekdayRangeTests = []struct {
	now    time.Time
	wd1    string
	wd2    string
	gmt    string
	result bool
}{
	{sundayUTC, "SUN", "", "", true},
	{mondayUTC, "MON", "", "", true},
	{tuesdayUTC, "TUE", "", "", true},
	{wednesdayUTC, "WED", "", "", true},
	{thursdayUTC, "THU", "", "", true},
	{fridayUTC, "FRI", "", "", true},
	{saturdayUTC, "SAT", "", "", true},
	{mondayUTC, "x", "", "", false},
	{mondayUTC, "x", "y", "", false},
	{mondayUTC, "x", "y", "z", false},
	{mondayUTC, "", "", "", false},
	{mondayUTC, "SUN", "", "", false},
	{mondayUTC, "SUN", "MON", "", true},
	{mondayUTC, "MON", "SUN", "", true},
	{mondayUTC.In(ny), "MON", "", "", false},
	{mondayUTC.In(ny), "MON", "GMT", "", true},
	{mondayUTC.In(ny), "MON", "", "GMT", true},
	{mondayUTC.In(ny), "SUN", "", "", true},
	{wednesdayUTC, "SUN", "WED", "", true},
	{wednesdayUTC, "MON", "WED", "", true},
	{wednesdayUTC, "WED", "SAT", "", true},
	{wednesdayUTC, "WED", "SUN", "", true},
	{wednesdayUTC, "SUN", "TUE", "", false},
	{wednesdayUTC, "MON", "TUE", "", false},
	{wednesdayUTC, "TUE", "SAT", "", true},
	{wednesdayUTC, "TUE", "SUN", "", false},
}

func TestWeekdayRange(t *testing.T) {
	defer func() {
		DefaultNower = &TimeNower{}
	}()
	for i, tt := range weekdayRangeTests {
		DefaultNower = &StaticNower{tt.now}
		result := WeekdayRange(tt.wd1, tt.wd2, tt.gmt)
		if result != tt.result {
			t.Errorf("Expecting test %d (%q, %q, %q) to return %v", i, tt.wd1, tt.wd2, tt.gmt, tt.result)
		}
	}
}

var dateRangeTests = []struct {
	now    time.Time
	args   []string
	result bool
}{
	{mondayUTC, []string{}, false},
	{mondayUTC, []string{"FOO"}, false},
	{mondayUTC, []string{"FOO", "BAR"}, false},
	{mondayUTC, []string{"1", "BAR"}, false},
	{mondayUTC, []string{"1"}, true},
	{mondayUTC, []string{"JAN"}, true},
	{mondayUTC, []string{"2018"}, true},
	{mondayUTC, []string{"2"}, false},
	{mondayUTC, []string{"FEB"}, false},
	{mondayUTC, []string{"2019"}, false},
	{mondayUTC.In(ny), []string{"1"}, false},
	{mondayUTC.In(ny), []string{"JAN"}, false},
	{mondayUTC.In(ny), []string{"2018"}, false},
	{mondayUTC.In(ny), []string{"1", "GMT"}, true},
	{mondayUTC.In(ny), []string{"JAN", "GMT"}, true},
	{mondayUTC.In(ny), []string{"2018", "GMT"}, true},
	{mondayUTC.In(ny), []string{"31"}, true},
	{mondayUTC.In(ny), []string{"DEC"}, true},
	{mondayUTC.In(ny), []string{"2017"}, true},
	{wednesdayUTC, []string{"3"}, true},
	{wednesdayUTC, []string{"3", "JAN"}, true},
	{wednesdayUTC, []string{"JAN", "3"}, true},
	{wednesdayUTC, []string{"JAN", "2018"}, true},
	{wednesdayUTC, []string{"2018", "JAN"}, true},
	{wednesdayUTC, []string{"2018", "JAN", "3"}, true},
	{wednesdayUTC, []string{"3", "JAN", "2018"}, true},
	{wednesdayUTC, []string{"JAN", "3", "2018"}, true},
	{thursdayUTC, []string{"1", "3"}, false},
	{thursdayUTC, []string{"1", "4"}, true},
	{thursdayUTC, []string{"4", "31"}, true},
	{thursdayUTC, []string{"5", "31"}, false},
	{thursdayUTC, []string{"1", "JAN", "3", "JAN"}, false},
	{thursdayUTC, []string{"1", "JAN", "4", "JAN"}, true},
	{thursdayUTC, []string{"4", "JAN", "31", "JAN"}, true},
	{thursdayUTC, []string{"5", "JAN", "31", "JAN"}, false},
	{thursdayUTC, []string{"1", "JAN", "2018", "3", "JAN"}, false},
	{thursdayUTC, []string{"1", "JAN", "2018", "4", "JAN"}, true},
	{thursdayUTC, []string{"4", "JAN", "2018", "31", "JAN"}, true},
	{thursdayUTC, []string{"5", "JAN", "2018", "31", "JAN"}, false},
	{thursdayUTC, []string{"1", "JAN", "2018", "3", "JAN", "2018"}, false},
	{thursdayUTC, []string{"1", "JAN", "2018", "4", "JAN", "2018"}, true},
	{thursdayUTC, []string{"4", "JAN", "2018", "31", "JAN", "2018"}, true},
	{thursdayUTC, []string{"5", "JAN", "2018", "31", "JAN", "2018"}, false},
	{thursdayUTC, []string{"1", "JAN", "2017", "3", "JAN", "2018"}, false},
	{thursdayUTC, []string{"1", "JAN", "2017", "4", "JAN", "2018"}, true},
	{thursdayUTC, []string{"4", "JAN", "2017", "31", "JAN", "2018"}, true},
	{thursdayUTC, []string{"5", "JAN", "2017", "31", "JAN", "2018"}, true},
	{thursdayUTC, []string{"1", "JAN", "2017", "3", "JAN", "2016"}, false},
	{thursdayUTC, []string{"1", "JAN", "2017", "4", "JAN", "2016"}, false},
	{thursdayUTC, []string{"4", "JAN", "2017", "31", "JAN", "2016"}, false},
	{thursdayUTC, []string{"5", "JAN", "2017", "31", "JAN", "2016"}, false},
	{thursdayUTC, []string{"1", "JAN", "2018", "3", "JAN", "2016"}, false},
	{thursdayUTC, []string{"1", "JAN", "2018", "4", "JAN", "2016"}, false},
	{thursdayUTC, []string{"4", "JAN", "2018", "31", "JAN", "2016"}, true},
	{thursdayUTC, []string{"5", "JAN", "2018", "31", "JAN", "2016"}, true},
	{thursdayUTC, []string{"1", "JAN", "2016", "3", "JAN", "2018"}, false},
	{thursdayUTC, []string{"1", "JAN", "2016", "4", "JAN", "2018"}, true},
	{thursdayUTC, []string{"4", "JAN", "2016", "31", "JAN", "2018"}, true},
	{thursdayUTC, []string{"5", "JAN", "2016", "31", "JAN", "2018"}, true},
}

func TestDateRange(t *testing.T) {
	defer func() {
		DefaultNower = &TimeNower{}
	}()
	for i, tt := range dateRangeTests {
		DefaultNower = &StaticNower{tt.now}
		result := DateRange(tt.args)
		if result != tt.result {
			t.Errorf("Expecting test %d (%v, %v) to return %v", i, tt.now.Format(time.RFC3339Nano), tt.args, tt.result)
		}
	}

}

var timeRangeTests = []struct {
	now    time.Time
	args   []string
	result bool
}{
	{atTime(0, 0, 0), []string{}, false},
	{atTime(0, 0, 0), []string{"0"}, true},
	{atTime(0, 0, 0), []string{"0", "1"}, true},
	{atTime(0, 0, 0).In(ny), []string{"19"}, true},
	{atTime(0, 0, 0).In(ny), []string{"19", "20"}, true},
	{atTime(0, 0, 0).In(ny), []string{"19", "GMT"}, false},
	{atTime(0, 0, 0).In(ny), []string{"19", "20", "GMT"}, false},
	{atTime(0, 0, 0).In(ny), []string{"0", "GMT"}, true},
	{atTime(0, 0, 0).In(ny), []string{"0", "1", "GMT"}, true},
	{atTime(12, 0, 0).Add(-1), []string{"12"}, false},
	{atTime(12, 0, 0), []string{"12"}, true},
	{atTime(12, 30, 0), []string{"12"}, true},
	{atTime(13, 0, 0).Add(-1), []string{"12"}, true},
	{atTime(13, 0, 0), []string{"12"}, false},
	{atTime(12, 0, 0).Add(-1), []string{"12", "13"}, false},
	{atTime(12, 0, 0), []string{"12", "13"}, true},
	{atTime(12, 30, 0), []string{"12", "13"}, true},
	{atTime(13, 0, 0).Add(-1), []string{"12", "13"}, true},
	{atTime(13, 0, 0), []string{"12", "13"}, false},
	{atTime(12, 0, 0).Add(-1), []string{"13", "12"}, false},
	{atTime(12, 0, 0), []string{"13", "12"}, true},
	{atTime(12, 30, 0), []string{"13", "12"}, true},
	{atTime(13, 0, 0).Add(-1), []string{"13", "12"}, true},
	{atTime(13, 0, 0), []string{"13", "12"}, false},
	{atTime(8, 30, 0).Add(-1), []string{"8", "30", "17", "0"}, false},
	{atTime(8, 30, 0), []string{"8", "30", "17", "0"}, true},
	{atTime(13, 15, 0), []string{"8", "30", "17", "0"}, true},
	{atTime(17, 0, 0).Add(-1), []string{"8", "30", "17", "0"}, true},
	{atTime(17, 0, 0), []string{"8", "30", "17", "0"}, false},
	{atTime(0, 0, 0).Add(-1), []string{"0", "0", "0", "0", "0", "30"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "0", "0", "0", "30"}, true},
	{atTime(0, 0, 15), []string{"0", "0", "0", "0", "0", "30"}, true},
	{atTime(0, 0, 30).Add(-1), []string{"0", "0", "0", "0", "0", "30"}, true},
	{atTime(0, 0, 30), []string{"0", "0", "0", "0", "0", "30"}, false},
	{atTime(0, 0, 0).Add(-1), []string{"0", "0", "30", "0", "0", "0"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "30", "0", "0", "0"}, true},
	{atTime(0, 0, 15), []string{"0", "0", "30", "0", "0", "0"}, true},
	{atTime(0, 0, 30).Add(-1), []string{"0", "0", "30", "0", "0", "0"}, true},
	{atTime(0, 0, 30), []string{"0", "0", "30", "0", "0", "0"}, false},
	{atTime(0, 0, 15).In(ny), []string{"0", "0", "0", "0", "0", "30"}, false},
	{atTime(0, 0, 15).In(ny), []string{"0", "0", "0", "0", "0", "30", "GMT"}, true},
	{atTime(0, 0, 0), []string{"0"}, true},
	{atTime(0, 0, 0), []string{"x"}, false},
	{atTime(0, 0, 0), []string{"0", "1"}, true},
	{atTime(0, 0, 0), []string{"x", "0"}, false},
	{atTime(0, 0, 0), []string{"0", "x"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "0", "1", "1", "1"}, true},
	{atTime(0, 0, 0), []string{"x", "0", "0", "1", "1", "1"}, false},
	{atTime(0, 0, 0), []string{"0", "x", "0", "1", "1", "1"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "x", "1", "1", "1"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "0", "x", "1", "1"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "0", "1", "x", "1"}, false},
	{atTime(0, 0, 0), []string{"0", "0", "0", "1", "1", "x"}, false},
	{atTime(1, 2, 3), []string{"1"}, true},
	{atTime(1, 2, 3), []string{"1", "2"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "3"}, false},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "5"}, false},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "6"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "6", "7"}, false},
	{atNYTime(1, 2, 3), []string{"1", "2"}, true},
	{atNYTime(1, 2, 3), []string{"1", "2", "3"}, false},
	{atNYTime(1, 2, 3), []string{"1", "2", "3", "4"}, true},
	{atNYTime(1, 2, 3), []string{"1", "2", "3", "4", "5"}, false},
	{atNYTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "6"}, true},
	{atNYTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "6", "7"}, false},
	{atTime(1, 2, 3), []string{"1", "GMT"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "GMT"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "3", "GMT"}, false},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "GMT"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "GMT"}, false},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "6", "GMT"}, true},
	{atTime(1, 2, 3), []string{"1", "2", "3", "4", "5", "6", "7", "GMT"}, false},
	{atTime(1, 2, 3).In(ny), []string{"1", "GMT"}, true},
	{atTime(1, 2, 3).In(ny), []string{"1", "2", "GMT"}, true},
	{atTime(1, 2, 3).In(ny), []string{"1", "2", "3", "GMT"}, false},
	{atTime(1, 2, 3).In(ny), []string{"1", "2", "3", "4", "GMT"}, true},
	{atTime(1, 2, 3).In(ny), []string{"1", "2", "3", "4", "5", "GMT"}, false},
	{atTime(1, 2, 3).In(ny), []string{"1", "2", "3", "4", "5", "6", "GMT"}, true},
	{atTime(1, 2, 3).In(ny), []string{"1", "2", "3", "4", "5", "6", "7", "GMT"}, false},
}

func atTime(h, m, s int) time.Time {
	return time.Date(2018, 1, 1, h, m, s, 0, time.UTC)
}

func atNYTime(h, m, s int) time.Time {
	return time.Date(2018, 1, 1, h, m, s, 0, ny)
}

func TestTimeRange(t *testing.T) {
	defer func() {
		DefaultNower = &TimeNower{}
	}()
	for i, tt := range timeRangeTests {
		DefaultNower = &StaticNower{tt.now}
		result := TimeRange(tt.args)
		if result != tt.result {
			t.Errorf("Expecting test %d (%v, %v) to return %v", i, tt.now.Format(time.RFC3339Nano), tt.args, tt.result)
		}
	}
}
