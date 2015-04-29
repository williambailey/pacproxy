package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewPacConn(t *testing.T) {
	n := time.Now().Nanosecond()
	p := NewPacConn(
		"example.com",
		time.Second,
		time.Second,
		time.Second,
	)
	require.Equal(t, "example.com", p.Address())
	require.InDelta(t, n, p.Created().Nanosecond(), float64(time.Millisecond))
	require.InDelta(t, n, p.Updated().Nanosecond(), float64(time.Millisecond))
	require.Equal(t, p.Created(), p.Updated())
}
