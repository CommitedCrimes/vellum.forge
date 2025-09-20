package version

import (
	"testing"

	"vellum.forge/internal/assert"
)

func TestGet(t *testing.T) {
	t.Run("Returns a non-empty string", func(t *testing.T) {
		version := Get()
		assert.True(t, version != "")
	})
}
