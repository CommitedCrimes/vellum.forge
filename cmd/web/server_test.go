package main

import (
	"net"

	"testing"

	"vellum.forge/internal/assert"
)

func TestServerConfiguration(t *testing.T) {
	t.Run("Default timeouts are reasonable", func(t *testing.T) {
		assert.True(t, defaultIdleTimeout > 0)
		assert.True(t, defaultReadTimeout > 0)
		assert.True(t, defaultWriteTimeout > defaultReadTimeout)

		if defaultShutdownPeriod <= defaultWriteTimeout {
			t.Errorf("default shutdown period %s must be greater than default write timeout %s", defaultShutdownPeriod, defaultWriteTimeout)
		}
	})
}

func TestServeHTTP(t *testing.T) {

	t.Run("Invalid port configuration causes an error", func(t *testing.T) {
		app := newTestApplication(t)
		app.config.httpPort = -1

		err := app.serveHTTP()
		assert.NotNil(t, err)
	})
}

func getFreePort(t *testing.T) int {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}
