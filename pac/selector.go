package pac

// FirstItemSelector simply selects the first item or returns DirectProxy
type FirstItemSelector struct{}

func (s *FirstItemSelector) SelectProxy(from Proxies) Proxy {
	if len(from) < 1 {
		return DirectProxy
	}
	return from[0]
}
