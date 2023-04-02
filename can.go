package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Can struct {
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

func (C *Can) Run() error {
	if debugOn {
		debug.SetOutput(os.Stderr)
	}

	if len(C.Input) == 0 {
		return fmt.Errorf("missing input")
	}

	if err := C.loadkey(); err != nil {
		return err
	}

	if C.API.URL == nil {
		C.API.URL, _ = url.Parse("https://api.openai.com")
	}

	// select action
	var cmd Command
	switch {
	case C.Src != "":
		c := NewEdits()
		if err := c.SetInput(C.Src); err != nil {
			return err
		}
		c.UpdateSrc = c.UpdateSrc
		c.Instruction = C.Input
		cmd = c

	default:
		c := NewChat()
		c.Content = C.Input
		c.SystemContent = C.SysContent
		cmd = c
	}

	// execute action
	r := cmd.MakeRequest()
	r.Header.Set("authorization", "Bearer "+C.API.Key)
	r.URL.Host = C.API.URL.Host
	r.URL.Scheme = C.API.URL.Scheme

	body, err := sendRequest(r)
	if err != nil {
		return err
	}

	return cmd.HandleResponse(body)
}

func (C *Can) loadkey() error {
	if len(C.API.Key) > 0 {
		return nil
	}
	if C.API.KeyFile == "" {
		return nil
	}
	data, err := os.ReadFile(C.API.KeyFile)
	if err != nil {
		return err
	}
	C.API.Key = string(bytes.TrimSpace(data))
	return nil
}

type Command interface {
	MakeRequest() *http.Request
	HandleResponse(io.Reader) error
}
