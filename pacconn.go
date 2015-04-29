package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type PacConn struct {
	mutex       *sync.RWMutex
	address     string
	created     time.Time
	updated     time.Time
	dialRetry   time.Duration
	dialTimeout time.Duration
	idleTimeout time.Duration
	err         error
	conn        net.Conn
}

func NewPacConn(address string, dialTimeout, idleTimeout, dialRetry time.Duration) *PacConn {
	n := time.Now()
	return &PacConn{
		mutex:       &sync.RWMutex{},
		address:     address,
		created:     n,
		updated:     n,
		dialRetry:   dialRetry,
		dialTimeout: dialTimeout,
		idleTimeout: idleTimeout,
	}
}

func (c *PacConn) Address() string {
	return c.address
}

func (c *PacConn) Created() time.Time {
	return c.created
}

func (c *PacConn) Updated() time.Time {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.updated
}

func (c *PacConn) Error() error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.dialErr()
}

func (c *PacConn) IsActive() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.conn != nil {
		return true
	}
	now := time.Now()
	if c.err != nil && c.updated.Add(c.dialRetry).After(now) {
		return false
	}
	c.conn, c.err = net.DialTimeout("tcp", c.address, c.dialTimeout)
	c.updated = now
	if c.err == nil {
		go func() {
			time.Sleep(c.idleTimeout)
			c.mutex.Lock()
			defer c.mutex.Unlock()
			if c.conn != nil {
				c.conn.Close()
				c.conn = nil
			}
		}()
		return true
	}
	return false
}

func (c *PacConn) BlacklistDuration() time.Duration {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	now := time.Now()
	expire := c.updated.Add(c.dialRetry)
	duration := expire.Sub(now)
	if duration < 0 {
		return 0
	}
	return duration
}

func (c *PacConn) Dial() (net.Conn, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	conn := c.conn
	if conn != nil {
		c.conn = nil
		return conn, nil
	}
	now := time.Now()
	if c.err != nil && c.updated.Add(c.dialRetry).After(now) {
		return nil, c.dialErr()
	}
	conn, c.err = net.DialTimeout("tcp", c.address, c.dialTimeout)
	c.conn = nil
	c.updated = now
	if c.err != nil {
		return nil, c.dialErr()
	}
	return conn, nil
}

func (c *PacConn) dialErr() error {
	if c.err == nil {
		return nil
	}
	now := time.Now()
	expire := c.updated.Add(c.dialRetry)
	return fmt.Errorf("Connection to %q is currently blacklisted for %s: %s", c.address, expire.Sub(now), c.err)
}

type PacConnService struct {
	mutex   *sync.RWMutex
	pacConn map[string]*PacConn
}

func NewPacConnService() *PacConnService {
	return &PacConnService{
		mutex:   &sync.RWMutex{},
		pacConn: make(map[string]*PacConn, 0),
	}
}

func (s *PacConnService) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.pacConn = make(map[string]*PacConn, 0)
}

func (s *PacConnService) KnownProxies() map[string]*PacConn {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	m := make(map[string]*PacConn, len(s.pacConn))
	for k, v := range s.pacConn {
		m[k] = v
	}
	return m
}

func (s *PacConnService) IsKnownProxy(address string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	pc := s.pacConn[address]
	return pc != nil
}

func (s *PacConnService) Conn(address string) *PacConn {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	pc := s.pacConn[address]
	if pc == nil {
		pc = NewPacConn(address, time.Second*5, time.Second*10, time.Minute*5)
		s.pacConn[address] = pc
	}
	return pc
}
