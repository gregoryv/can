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
	"fmt"
	"io"
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


func Test_chat(t *testing.T) {
	c := newChat()

	if r := c.MakeRequest(); r == nil {
		t.Error("nil request")
	}

	if err := c.HandleResponse(strings.NewReader("{}")); err == nil {
		t.Error("empty result should fail")
	}

	if err := c.HandleResponse(strings.NewReader(validCompletionsResponse)); err != nil {
		t.Error(err)
	}

	// invalid json
	if err := c.HandleResponse(strings.NewReader("{")); err == nil {
		t.Error("expect error on invalid json")
	}

}

const validCompletionsResponse = `{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "\n\nHello there, how may I assist you today?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 9,
    "completion_tokens": 12,
    "total_tokens": 21
  }
}`


func Test_edits_MakeRequest(t *testing.T) {
	c := newEdits()

	if r := c.MakeRequest(); r == nil {
		t.Error("nil request")
	}
}

func Test_edits_SetInput(t *testing.T) {
	c := newEdits()
	// plain text
	c.SetInput("some text that is not a file")

	// from file
	dst := filepath.Join(os.TempDir(), "edits.txt")
	defer func() {
		os.Chmod(dst, 0600)
		os.RemoveAll(dst)
	}()
	os.WriteFile(dst, []byte(""), 0644)
	if err := c.SetInput(dst); err != nil {
		t.Error(err)
	}

	// bad permission
	os.Chmod(dst, 0100)
	if c.SetInput(dst) == nil {
		t.Error("expect error on inadequate read permission")
	}
}

func Test_edits_HandleResponse(t *testing.T) {
	c := newEdits()

	if err := c.HandleResponse(strings.NewReader("{}")); err == nil {
		t.Error("empty result should fail")
	}

	if err := c.HandleResponse(strings.NewReader(validEditsResponse)); err != nil {
		t.Error(err)
	}

	// invalid json
	if err := c.HandleResponse(strings.NewReader("{")); err == nil {
		t.Error("expect error on invalid json")
	}

	// check result is written to file
	dst := filepath.Join(os.TempDir(), "edits.txt")
	defer func() {
		os.Chmod(dst, 0600)
		os.RemoveAll(dst)
	}()
	os.WriteFile(dst, []byte(""), 0644)
	c.SetInput(dst)
	c.UpdateSrc = true
	if err := c.HandleResponse(strings.NewReader(validEditsResponse)); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dst)
	if v := string(got); v != "word" {
		t.Errorf("got %q", v)
	}

	os.Chmod(dst, 0500)
	if err := c.HandleResponse(strings.NewReader(validEditsResponse)); err == nil {
		t.Fatal("expect error on inadequate write permission")
	}
}

// from https://platform.openai.com/docs/api-reference/edits/create
const validEditsResponse = `{
  "object": "edit",
  "created": 1589478378,
  "choices": [
    {
      "text": "word",
      "index": 0
    }
  ],
  "usage": {
    "prompt_tokens": 25,
    "completion_tokens": 32,
    "total_tokens": 57
  }
}
`


func Test_should(t *testing.T) {
	if got := should([]byte("in"), io.EOF); string(got) != "in" {
		t.Fail()
	}
}

func serve(v string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, v)
	})
}
