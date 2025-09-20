package main

import (
	"net/http"
	"testing"

	"vellum.forge/internal/assert"
)

func TestHome(t *testing.T) {
	t.Run("GET renders the home page", func(t *testing.T) {
		app := newTestApplication(t)

		req := newTestRequest(t, http.MethodGet, "/")

		res := send(t, req, app.routes())
		assert.Equal(t, res.StatusCode, http.StatusOK)
		assert.True(t, containsPageTag(t, res.Body, "home"))
	})
}
