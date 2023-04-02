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
	API struct {
		*url.URL // host:port
		KeyFile  string
		Key      string
	}

	SysContent string
	UpdateSrc  bool
	Src        string // ie. file or block of text
	Input      string
}

func (s *System) Run() error {
	if len(s.Input) == 0 {
		return fmt.Errorf("missing input")
	}
	if err := s.loadkey(); err != nil {
		return err
	}
	if s.API.URL == nil {
		return fmt.Errorf("Can.Run: missing API.URL")
	}

	// select action
	var cmd Command
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
		c := NewChat()
		c.Content = s.Input
		c.SystemContent = s.SysContent
		cmd = c
	}

	// execute action
	r := cmd.MakeRequest()
	r.Header.Set("authorization", "Bearer "+s.API.Key)
	r.URL.Host = s.API.URL.Host
	r.URL.Scheme = s.API.URL.Scheme

	body, err := sendRequest(r)
	if err != nil {
		return err
	}

	return cmd.HandleResponse(body)
}

func (s *System) loadkey() error {
	if s.API.Key != "" {
		return nil
	}
	if s.API.KeyFile == "" {
		return nil
	}
	data, err := os.ReadFile(s.API.KeyFile)
	if err != nil {
		return err
	}
	s.API.Key = string(bytes.TrimSpace(data))
	return nil
}

type Command interface {
	MakeRequest() *http.Request
	HandleResponse(io.Reader) error
}
