package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
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
	c.API.Key = "secret"
	c.SysContent = "As a nice assistant."
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

func TestCan_loadkey(t *testing.T) {
	var c Can
	if err := c.loadkey(); err != nil {
		t.Error(err)
	}

	dst := filepath.Join(os.TempDir(), "somefile")
	defer func() {
		os.Chmod(dst, 0500)
		os.RemoveAll(dst)
	}()

	_ = os.WriteFile(dst, []byte("secret"), 0400)
	c.API.KeyFile = dst
	if err := c.loadkey(); err != nil {
		t.Error(err)
	}

	// without read permission
	os.Chmod(dst, 0000)
	c.API.Key = "" // reset
	if err := c.loadkey(); err == nil {
		t.Error("expect error")
	}

}
