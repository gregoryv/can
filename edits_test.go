package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEdits(t *testing.T) {
	c := NewEdits()

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

	// check result is written to file
	dst := filepath.Join(os.TempDir(), "edits.txt")
	os.WriteFile(dst, []byte(""), 0644)
	c.Src = dst
	c.UpdateSrc = true
	if err := c.HandleResponse(strings.NewReader(valid)); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dst)
	if v := string(got); v != "word" {
		t.Errorf("got %q", v)
	}
}

// from https://platform.openai.com/docs/api-reference/edits/create
const valid = `{
  "object": "edit",
  "created": 1589478378,
  "choices": [
    {
      "text": "word",
      "index": 0
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 32,
    "total_tokens": 57
  }
}
`
