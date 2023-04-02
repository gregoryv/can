package can

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type System struct {
	api struct {
		*url.URL // host:port
		KeyFile  string
		Key      string
	}

	sysContent string
	UpdateSrc  bool
	Src        string // ie. file or block of text
	Input      string
}

func (s *System) SetAPIUrl(v *url.URL)   { s.api.URL = v }
func (s *System) SetAPIKey(v string)     { s.api.Key = v }
func (s *System) SetAPIKeyFile(v string) { s.api.KeyFile = v }
func (s *System) SetSysContent(v string) { s.sysContent = v }

func (s *System) Run() error {
	if len(s.Input) == 0 {
		return fmt.Errorf("missing input")
	}
	if err := s.loadkey(); err != nil {
		return err
	}
	if s.api.URL == nil {
		return fmt.Errorf("Can.Run: missing API.URL")
	}

	// select action
	var cmd command
	switch {
	case s.Src != "":
		c := newEdits()
		if err := c.SetInput(s.Src); err != nil {
			return err
		}
		c.UpdateSrc = c.UpdateSrc
		c.Instruction = s.Input
		cmd = c

	default:
		c := newChat()
		c.Content = s.Input
		c.SystemContent = s.sysContent
		cmd = c
	}

	// execute action
	r := cmd.MakeRequest()
	r.Header.Set("authorization", "Bearer "+s.api.Key)
	r.URL.Host = s.api.URL.Host
	r.URL.Scheme = s.api.URL.Scheme

	body, err := sendRequest(r)
	if err != nil {
		return err
	}

	return cmd.HandleResponse(body)
}

func (s *System) loadkey() error {
	if s.api.Key != "" {
		return nil
	}
	if s.api.KeyFile == "" {
		return nil
	}
	data, err := os.ReadFile(s.api.KeyFile)
	if err != nil {
		return err
	}
	s.api.Key = string(bytes.TrimSpace(data))
	return nil
}

type command interface {
	MakeRequest() *http.Request
	HandleResponse(io.Reader) error
}
