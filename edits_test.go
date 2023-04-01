package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEdits_MakeRequest(t *testing.T) {
	c := NewEdits()

	if r := c.MakeRequest(); r == nil {
		t.Error("nil request")
	}
}

func TestEdits_SetSrc(t *testing.T) {
	c := NewEdits()
	// plain text
	c.SetSrc("some text that is not a file")

	// from file
	dst := filepath.Join(os.TempDir(), "edits.txt")
	defer func() {
		os.Chmod(dst, 0600)
		os.RemoveAll(dst)
	}()
	os.WriteFile(dst, []byte(""), 0644)
	if err := c.SetSrc(dst); err != nil {
		t.Error(err)
	}

	// bad permission
	os.Chmod(dst, 0100)
	if c.SetSrc(dst) == nil {
		t.Error("expect error on inadequate read permission")
	}
}

func TestEdits_HandleResponse(t *testing.T) {
	c := NewEdits()

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
	defer func() {
		os.Chmod(dst, 0600)
		os.RemoveAll(dst)
	}()
	os.WriteFile(dst, []byte(""), 0644)
	c.SetSrc(dst)
	c.UpdateSrc = true
	if err := c.HandleResponse(strings.NewReader(valid)); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dst)
	if v := string(got); v != "word" {
		t.Errorf("got %q", v)
	}

	os.Chmod(dst, 0500)
	if err := c.HandleResponse(strings.NewReader(valid)); err == nil {
		t.Fatal("expect error on inadequate write permission")
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
