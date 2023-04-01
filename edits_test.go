package main

import "testing"

func TestEdits(t *testing.T) {
	c := NewEdits()

	if _, err := c.makeRequest(); err != nil {
		t.Error(err)
	}
}
