package main

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"
)

func TestChat(t *testing.T) {
	c := NewChat()

	if r := c.MakeRequest(); r == nil {
		t.Error("nil request")
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

func Test_readClose(t *testing.T) {
	debugOn = true // global
	in := `{"name": "carl"}`
	var buf bytes.Buffer
	debug.SetOutput(&buf)
	readClose(ioutil.NopCloser(strings.NewReader(in)))
	if buf.Len() == 0 {
		t.Fail()
	}
}
