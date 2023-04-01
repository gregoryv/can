package main

import (
	"strings"
	"testing"
)

func TestEdits(t *testing.T) {
	c := NewEdits()

	if _, err := c.makeRequest(); err != nil {
		t.Error(err)
	}

	empty := strings.NewReader("{}")
	if err := c.handleResponse(empty); err == nil {
		t.Error("empty result should result in error")
	}
}
