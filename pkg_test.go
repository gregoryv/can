package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"
)

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

func Test_should(t *testing.T) {
	if got := should([]byte("in"), io.EOF); string(got) != "in" {
		t.Fail()
	}
}
