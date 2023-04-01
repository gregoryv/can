package main

import (
	"strings"
	"testing"
)

func TestChat(t *testing.T) {
	c := NewChat()

	if _, err := c.makeRequest(); err != nil {
		t.Error(err)
	}

	empty := strings.NewReader("{}")
	if err := c.handleResponse(empty); err == nil {
		t.Error("empty result should result in error")
	}
}
