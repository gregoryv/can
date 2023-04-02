package main

import (
	"testing"
)

func TestCan_Run(t *testing.T) {
	var c Can
	if err := c.Run(); err == nil {
		t.Error("empty can should fail")
	}
}
