package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

func Test_sendRequest(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		s := httptest.NewServer(serve("{}", 200))
		defer s.Close()

		r, _ := http.NewRequest("GET", s.URL, http.NoBody)
		body, _ := sendRequest(r)
		if body == nil {
			t.Error("unexpected empty body")
		}
	})

	t.Run("bad request", func(t *testing.T) {
		s := httptest.NewServer(serve("{}", 400))
		defer s.Close()

		r, _ := http.NewRequest("GET", s.URL, http.NoBody)
		if _, err := sendRequest(r); err == nil {
			t.Error("unexpected error")
		}
	})

	t.Run("server down", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "http://localhost:12345", http.NoBody)
		if _, err := sendRequest(r); err == nil {
			t.Error("bad request was ok")
		}
	})

}

func serve(v string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, v)
	})
}
