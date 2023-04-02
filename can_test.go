package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestCan_RunIssues(t *testing.T) {
	cases := []struct {
		txt string
		*Can
	}{
		{"empty", &Can{}},
		{"unreadable key file",
			func() *Can {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("secret"), 0000)
				var c Can
				c.Input = "some text"
				c.API.KeyFile = dst
				return &c
			}(),
		},
		{"missing API.URL",
			func() *Can {
				var c Can
				c.Input = "some text"
				c.API.Key = "secret"
				return &c
			}(),
		},
		{"unreadable src file",
			func() *Can {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("data"), 0000)
				var c Can
				c.Input = "some text"
				c.API.Key = "secret"
				c.API.URL, _ = url.Parse("http://example.com")
				c.Src = dst
				return &c
			}(),
		},
		{"unreadable src file",
			func() *Can {
				var c Can
				c.Src = "2 apples, 3 oranges"
				c.Input = "count fruits"
				c.API.Key = "secret"
				c.API.URL, _ = url.Parse("http://localhost:12345") // no such host
				return &c
			}(),
		},
	}
	for _, c := range cases {
		if err := c.Can.Run(); err == nil {
			t.Error(c.txt, err)
		}
	}
}

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

	dst := filepath.Join(t.TempDir(), "somefile")
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
