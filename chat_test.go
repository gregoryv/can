package main

import (
	"strings"
	"testing"
)

func TestChat(t *testing.T) {
	c := NewChat()
	_, err := c.MakeRequest()
	if err != nil {
		t.Error(err)
	}

	empty := strings.NewReader("{}")
	if err := c.HandleResponse(empty); err == nil {
		t.Error("empty result should result in error")
	}
}
