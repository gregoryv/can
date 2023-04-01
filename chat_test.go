package main

import "testing"

func TestChat(t *testing.T) {
	c := NewChat()

	if _, err := c.makeRequest(); err != nil {
		t.Error(err)
	}
}
