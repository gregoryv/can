package can

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestSystem_RunIssues(t *testing.T) {
	cases := []struct {
		txt string
		*System
	}{
		{"empty", &System{}},
		{"unreadable key file",
			func() *System {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("secret"), 0000)
				var c System
				c.Input = "some text"
				c.SetAPIKeyFile(dst)
				return &c
			}(),
		},
		{"missing API.URL",
			func() *System {
				var c System
				c.Input = "some text"
				c.SetAPIKey("secret")
				return &c
			}(),
		},
		{"unreadable src file",
			func() *System {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("data"), 0000)
				var c System
				c.Input = "some text"
				c.api.Key = "secret"
				u, _ := url.Parse("http://example.com")
				c.SetAPIUrl(u)
				c.SetSrc(dst)
				return &c
			}(),
		},
		{"unreadable src file",
			func() *System {
				var c System
				c.SetSrc("2 apples, 3 oranges")
				c.Input = "count fruits"
				c.api.Key = "secret"
				c.api.URL, _ = url.Parse("http://localhost:12345") // no such host
				return &c
			}(),
		},
	}
	for _, c := range cases {
		if err := c.System.Run(); err == nil {
			t.Error(c.txt, err)
		}
	}
}

func TestSystem_Run(t *testing.T) {
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

	var c System
	c.api.URL, _ = url.Parse(srv.URL)
	c.api.Key = "secret"
	c.SetSysContent("As a nice assistant.")
	c.Input = "Hello!"
	if err := c.Run(); err != nil {
		t.Error(err)
	}

	c.SetSrc("hallo warld")
	c.Input = "fix spelling"
	if err := c.Run(); err != nil {
		t.Error(err)
	}
}

func TestSystem_loadkey(t *testing.T) {
	var c System
	if err := c.loadkey(); err != nil {
		t.Error(err)
	}

	dst := filepath.Join(t.TempDir(), "somefile")
	_ = os.WriteFile(dst, []byte("secret"), 0400)
	c.api.KeyFile = dst
	if err := c.loadkey(); err != nil {
		t.Error(err)
	}

	// without read permission
	os.Chmod(dst, 0000)
	c.api.Key = "" // reset
	if err := c.loadkey(); err == nil {
		t.Error("expect error")
	}
}

func TestSystem_Settings(t *testing.T) {
	var s System
	s.SetUpdateSrc(true)
	s.SetUpdateSrc(false)
}
