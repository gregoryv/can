package can

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
				c.input = "some text"
				c.SetAPIKeyFile(dst)
				return &c
			}(),
		},
		{"missing API.URL",
			func() *System {
				var c System
				c.input = "some text"
				c.SetAPIKey("secret")
				return &c
			}(),
		},
		{"unreadable src file",
			func() *System {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("data"), 0000)
				var c System
				c.input = "some text"
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
				c.input = "count fruits"
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
	c.input = "Hello!"
	if err := c.Run(); err != nil {
		t.Error(err)
	}

	c.SetSrc("hallo warld")
	c.SetInput("fix spelling")
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

func TestSystem_sendRequest(t *testing.T) {
	var sys System
	t.Run("ok", func(t *testing.T) {
		s := httptest.NewServer(serve("{}", 200))
		defer s.Close()

		r, _ := http.NewRequest("GET", s.URL, http.NoBody)
		body, _ := sys.sendRequest(r)
		if body == nil {
			t.Error("unexpected empty body")
		}
	})

	t.Run("bad request", func(t *testing.T) {
		s := httptest.NewServer(serve("{}", 400))
		defer s.Close()

		r, _ := http.NewRequest("GET", s.URL, http.NoBody)
		if _, err := sys.sendRequest(r); err == nil {
			t.Error("unexpected error")
		}
	})

	t.Run("server down", func(t *testing.T) {
		r, _ := http.NewRequest("GET", "http://localhost:12345", http.NoBody)
		if _, err := sys.sendRequest(r); err == nil {
			t.Error("bad request was ok")
		}
	})

}

func Test_readClose(t *testing.T) {
	debugOn = true // global todo remove
	in := `{"name": "carl"}`
	var buf bytes.Buffer
	debug.SetOutput(&buf)
	var s System
	s.SetDebugOn(true)
	s.readClose(ioutil.NopCloser(strings.NewReader(in)))
	if buf.Len() == 0 {
		t.Fail()
	}
}
