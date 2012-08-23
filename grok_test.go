package grok

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil && g.g == nil {
		t.Fatal("Failed to initialize grok")
	}
}
