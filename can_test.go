package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCan_Run(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/edits":
			serve(validEditsResponse, 200).ServeHTTP(w, r)

		case "/v1/chat/completions":
			serve(validCompletionsResponse, 200).ServeHTTP(w, r)
		default:
			t.Fatal("oups", r.URL.Path)
		}
	}))
	defer srv.Close()

	var c Can
	c.API.URL, _ = url.Parse(srv.URL)
	c.Input = "Hello!"
	if err := c.Run(); err != nil {
		t.Error(err)
	}

	c.Src = "hallo warld"
	c.Input = "fix spelling"
	if err := c.Run(); err != nil {
		t.Error(err)
	}
}
