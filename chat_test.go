package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestChat(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "{}")
	}))
	defer s.Close()
	u, _ := url.Parse(s.URL)

	c := NewChat()

	r, err := c.MakeRequest()
	if err != nil {
		t.Error(err)
	}
	r.URL.Scheme = u.Scheme
	r.URL.Host = u.Host

	body, err := sendRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.HandleResponse(body); err == nil {
		t.Error("empty result should result in error")
	}
}
