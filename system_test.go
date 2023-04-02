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
		{"empty", NewSystem()},
		{"unreadable key file",
			func() *System {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("secret"), 0000)
				s := NewSystem()
				s.input = "some text"
				s.SetAPIKeyFile(dst)
				return s
			}(),
		},
		{"missing API.URL",
			func() *System {
				s := NewSystem()
				s.input = "some text"
				s.SetAPIKey("secret")
				return s
			}(),
		},
		{"unreadable src file",
			func() *System {
				dst := filepath.Join(t.TempDir(), "somefile")
				_ = os.WriteFile(dst, []byte("data"), 0000)
				s := NewSystem()
				s.input = "some text"
				s.api.Key = "secret"
				u, _ := url.Parse("http://example.com")
				s.SetAPIUrl(u)
				s.SetSrc(dst)
				return s
			}(),
		},
		{"unreadable src file",
			func() *System {
				s := NewSystem()
				s.SetSrc("2 apples, 3 oranges")
				s.input = "count fruits"
				s.api.Key = "secret"
				s.api.URL, _ = url.Parse("http://localhost:12345") // no such host
				return s
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

	s := NewSystem()
	s.api.URL, _ = url.Parse(srv.URL)
	s.api.Key = "secret"
	s.SetSysContent("As a nice assistant.")
	s.input = "Hello!"
	if err := s.Run(); err != nil {
		t.Error(err)
	}

	s.SetSrc("hallo warld")
	s.SetInput("fix spelling")
	if err := s.Run(); err != nil {
		t.Error(err)
	}
}

func TestSystem_loadkey(t *testing.T) {
	s := NewSystem()
	if err := s.loadkey(); err != nil {
		t.Error(err)
	}

	dst := filepath.Join(t.TempDir(), "somefile")
	_ = os.WriteFile(dst, []byte("secret"), 0400)
	s.SetAPIKeyFile(dst)
	if err := s.loadkey(); err != nil {
		t.Error(err)
	}

	// without read permission
	os.Chmod(dst, 0000)
	s.SetAPIKey("") // reset
	if err := s.loadkey(); err == nil {
		t.Error("expect error")
	}
}

func TestSystem_Settings(t *testing.T) {
	s := NewSystem()
	s.SetUpdateSrc(true)
	s.SetUpdateSrc(false)
	s.SetDebugOutput(nil)
}

func TestSystem_sendRequest(t *testing.T) {
	sys := NewSystem()
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
	var buf bytes.Buffer
	s := NewSystem()
	s.SetDebugOutput(&buf)
	in := `{"name": "carl"}`
	s.readClose(ioutil.NopCloser(strings.NewReader(in)))
	if buf.Len() == 0 {
		t.Fail()
	}
}
