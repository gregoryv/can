package main

import (
	"strings"
	"testing"
)

func TestChat(t *testing.T) {
	c := NewChat()

	if _, err := c.MakeRequest(); err != nil {
		t.Error(err)
	}

	if err := c.HandleResponse(strings.NewReader("{}")); err == nil {
		t.Error("empty result should fail")
	}

	if err := c.HandleResponse(strings.NewReader(valid)); err != nil {
		t.Error(err)
	}

	// invalid json
	if err := c.HandleResponse(strings.NewReader("{")); err == nil {
		t.Error("expect error on invalid json")
	}

}
