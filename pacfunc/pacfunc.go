package pacfunc

import (
	"encoding/binary"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// DefaultNower thats used by these functions to get the currant time
	DefaultNower Nower

	weekday = map[string]time.Weekday{
		"SUN": time.Sunday,
		"MON": time.Monday,
		"TUE": time.Tuesday,
		"WED": time.Wednesday,
		"THU": time.Thursday,
		"FRI": time.Friday,
		"SAT": time.Saturday,
	}
	month = map[string]time.Month{
		"JAN": time.January,
		"FEB": time.February,
		"MAR": time.March,
		"APR": time.April,
		"MAY": time.May,
		"JUN": time.June,
		"JUL": time.July,
		"AUG": time.August,
		"SEP": time.September,
		"OCT": time.October,
		"NOV": time.November,
		"DEC": time.December,
	}
)

func init() {
	DefaultNower = &TimeNower{}
}

// Nower is responsible for returning the current time
type Nower interface {
	Now() time.Time
}

// TimeNower implements Nower using the time package.
type TimeNower struct {
	static *time.Time
}

func (t TimeNower) Now() time.Time {
	if t.static != nil {
		return *t.static
	}
	return time.Now()
}

// StaticNower implements Nower with a static value
type StaticNower struct {
	now time.Time
}

func (s StaticNower) Now() time.Time {
	return s.now
}

// ConvertAddr converts an IPv4 dotted decimal IP address or an IPv6 IP address to an integer
func ConvertAddr(ipaddr string) uint32 {
	ip := net.ParseIP(ipaddr)
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

// DNSDomainIs evaluates hostnames and returns true if hostnames match.
//
// Used mainly to match and exception individual hostnames.
func DNSDomainIs(host, domain string) bool {
	if len(host) < len(domain) {
		return false
	}
	return strings.HasSuffix(host, domain)
}

// ShExpMatch will attempt to match hostname or URL to a specified shell expression, and returns true if matched.
func ShExpMatch(str, shexp string) bool {
	shexp = strings.Replace(shexp, ".", "\\.", -1)
	shexp = strings.Replace(shexp, "?", ".?", -1)
	shexp = strings.Replace(shexp, "*", ".*", -1)
	matched, err := regexp.MatchString("^"+shexp+"$", str)
	return err == nil && matched
}

// IsInNet evaluates the IP address of a hostname, and if within a specified
// subnet returns true. If a hostname is passed the function will resolve the
// hostname to an IP address.
func IsInNet(host, netip, netmask string) bool {
	if len(host) == 0 {
		return false
	}
	address, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return false
	}
	net := net.IPNet{
		IP:   net.ParseIP(netip),
		Mask: net.IPMask(net.ParseIP(netmask)),
	}
	return net.Contains(address.IP)
}

// MyIPAddress returns the IP address of the host machine.
func MyIPAddress() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "127.0.0.1"
	}
	return DNSResolve(hostname)
}

// DNSResolve returns the IP address of the host.
func DNSResolve(host string) string {
	address, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return ""
	}
	return address.String()
}

// IsPlainHostName will return true if the hostname contains no dots, e.g. http://intranet
//
// Useful when applying exceptions for internal websites, e.g. may not require
// resolution of a hostname to IP address to determine if local.
func IsPlainHostName(host string) bool {
	return strings.Index(host, ".") == -1
}

// LocalHostOrDomainIs evaluates hostname and only returns true if exact
// hostname match is found.
func LocalHostOrDomainIs(host, hostdom string) bool {
	if host == hostdom {
		return true
	}
	return strings.LastIndex(hostdom, host+".") == 0
}

// IsResolvable attempts to resolve a hostname to an IP address and returns
// true if successful.
func IsResolvable(host string) bool {
	if len(host) == 0 {
		return false
	}
	if _, err := net.ResolveIPAddr("ip", host); err != nil {
		return false
	}
	return true
}

// DNSDomainLevels returns the number of DNS domain levels (number of dots)
// in the hostname. Can be used to exception internal websites which use short
// DNS names, e.g. http://intranet
func DNSDomainLevels(host string) int {
	return strings.Count(host, ".")
}

// WeekdayRange return true if the current date is during that period
//
// Only the first parameter is mandatory. Either the second, the third, or
// both may be left out.
//
// If only one parameter is present, the function returns a value of true
// on the weekday that the parameter represents. If the string "GMT" is
// specified as a second parameter, times are taken to be in GMT. Otherwise,
// they are assumed to be in the local timezone.
//
// If both wd1 and wd1 are defined, the condition is true if the current
// weekday is in between those two ordered weekdays. Bounds are inclusive,
// but the bounds are ordered. If the "GMT" parameter is specified, times
// are taken to be in GMT. Otherwise, the local timezone is used.
func WeekdayRange(wd1, wd2, gmt string) bool {
	wd1 = strings.ToUpper(wd1)
	wd2 = strings.ToUpper(wd2)
	gmt = strings.ToUpper(gmt)
	if wd2 == "GMT" {
		wd2 = ""
		gmt = "GMT"
	}
	if wd2 == "" {
		wd2 = wd1
	}
	now := DefaultNower.Now()
	if gmt == "GMT" {
		now = now.UTC()
	}
	today := now.Weekday()
	var (
		ok       = true
		weekday1 time.Weekday
		weekday2 time.Weekday
	)
	if weekday1, ok = weekday[wd1]; !ok {
		return false
	}
	if weekday2, ok = weekday[wd2]; !ok {
		return false
	}
	if weekday1 == weekday2 && weekday1 == today {
		return true
	}
	if weekday1 > weekday2 {
		weekday2, weekday1 = weekday1, weekday2
	}
	return (weekday1 <= today) && (today <= weekday2)
}

// DateRange return true during (or between) the specified date(s).
//
// (<day1>, <month1>, <year1>, <day2>, <month2>, <year2>, <gmt>)
func DateRange(args []string) bool {
	getMonth := func(name string) time.Month {
		month, _ := month[name]
		return month
	}
	argc := len(args)
	if argc < 1 {
		return false
	}
	for k, v := range args {
		args[k] = strings.ToUpper(v)
	}
	now := DefaultNower.Now()
	isGMT := args[argc-1] == "GMT"
	if isGMT {
		argc--
		now = now.UTC()
	}
	if argc == 1 {
		tmp, err := strconv.Atoi(args[0])
		if err != nil {
			return now.Month() == getMonth(args[0])
		} else if tmp <= 31 {
			return now.Day() == tmp
		} else {
			return now.Year() == tmp
		}
	}
	date1 := now
	for i := 0; i < argc/2; i++ {
		tmp, err := strconv.Atoi(args[i])
		if err != nil {
			m := getMonth(args[i])
			if m == 0 {
				return false
			}
			date1 = time.Date(
				date1.Year(),
				m,
				date1.Day(),
				date1.Hour(),
				date1.Minute(),
				date1.Second(),
				date1.Nanosecond(),
				date1.Location(),
			)
		} else if tmp <= 31 {
			date1 = time.Date(
				date1.Year(),
				date1.Month(),
				tmp,
				date1.Hour(),
				date1.Minute(),
				date1.Second(),
				date1.Nanosecond(),
				date1.Location(),
			)
		} else {
			date1 = time.Date(
				tmp,
				date1.Month(),
				date1.Day(),
				date1.Hour(),
				date1.Minute(),
				date1.Second(),
				date1.Nanosecond(),
				date1.Location(),
			)
		}
	}
	date2 := now
	for i := argc / 2; i < argc; i++ {
		tmp, err := strconv.Atoi(args[i])
		if err != nil {
			m := getMonth(args[i])
			if m == 0 {
				return false
			}
			date2 = time.Date(
				date2.Year(),
				m,
				date2.Day(),
				date2.Hour(),
				date2.Minute(),
				date2.Second(),
				date2.Nanosecond(),
				date2.Location(),
			)
		} else if tmp <= 31 {
			date2 = time.Date(
				date2.Year(),
				date2.Month(),
				tmp,
				date2.Hour(),
				date2.Minute(),
				date2.Second(),
				date2.Nanosecond(),
				date2.Location(),
			)
		} else {
			date2 = time.Date(
				tmp,
				date2.Month(),
				date2.Day(),
				date2.Hour(),
				date2.Minute(),
				date2.Second(),
				date2.Nanosecond(),
				date2.Location(),
			)
		}
	}
	var (
		nano  = now.UnixNano()
		nano1 = date1.UnixNano()
		nano2 = date2.UnixNano()
	)
	if nano2 < nano1 {
		nano1, nano2 = nano2, nano1
	}
	return nano1 <= nano && nano <= nano2
}

// TimeRange return true during (or between) the specified time(s).
//
// (<hour1>, <min1>, <sec1>, <hour2>, <min2>, <sec2>, <gmt>)
func TimeRange(args []string) bool {
	argc := len(args)
	if argc < 1 {
		return false
	}
	for k, v := range args {
		args[k] = strings.ToUpper(v)
	}
	now := DefaultNower.Now()
	isGMT := args[argc-1] == "GMT"
	if isGMT {
		argc--
		now = now.UTC()
	}
	date1 := now
	date2 := now
	switch argc {
	case 1:
		tmp, err := strconv.Atoi(args[0])
		if err != nil {
			return false
		}
		return now.Hour() == tmp
	case 2:
		tmp1, err := strconv.Atoi(args[0])
		if err != nil {
			return false
		}
		tmp2, err := strconv.Atoi(args[1])
		if err != nil {
			return false
		}
		if tmp2 < tmp1 {
			tmp1, tmp2 = tmp2, tmp1
		}
		return tmp1 <= now.Hour() && now.Hour() < tmp2
	case 6:
		s1, err := strconv.Atoi(args[2])
		if err != nil {
			return false
		}
		s2, err := strconv.Atoi(args[5])
		if err != nil {
			return false
		}
		date1 = time.Date(
			date1.Year(),
			date1.Month(),
			date1.Day(),
			date1.Hour(),
			date1.Minute(),
			s1,
			date1.Nanosecond(),
			date1.Location(),
		)
		date2 = time.Date(
			date2.Year(),
			date2.Month(),
			date2.Day(),
			date2.Hour(),
			date2.Minute(),
			s2,
			date2.Nanosecond(),
			date2.Location(),
		)
		fallthrough
	case 4:
		middle := argc / 2
		h1, err := strconv.Atoi(args[0])
		if err != nil {
			return false
		}
		m1, err := strconv.Atoi(args[1])
		if err != nil {
			return false
		}
		h2, err := strconv.Atoi(args[middle])
		if err != nil {
			return false
		}
		m2, err := strconv.Atoi(args[middle+1])
		if err != nil {
			return false
		}
		date1 = time.Date(
			date1.Year(),
			date1.Month(),
			date1.Day(),
			h1,
			m1,
			date1.Second(),
			date1.Nanosecond(),
			date1.Location(),
		)
		date2 = time.Date(
			date2.Year(),
			date2.Month(),
			date2.Day(),
			h2,
			m2,
			date2.Second(),
			date2.Nanosecond(),
			date2.Location(),
		)
		break
	default:
		return false
	}
	var (
		nano  = now.UnixNano()
		nano1 = date1.UnixNano()
		nano2 = date2.UnixNano()
	)
	if nano2 < nano1 {
		nano1, nano2 = nano2, nano1
	}
	return (nano1 <= nano) && (nano < nano2)
}
