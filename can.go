package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type Can struct {
	API struct {
		Host    string // host:port
		KeyFile string
		Key     string
	}

	SysContent string
	Src        string // ie. file or block of text
	Input      string
}

func (C *Can) Run() error {

	log.SetFlags(0)
	if debugOn {
		debug.SetOutput(os.Stderr)
	}

	if len(C.Input) == 0 {
		return fmt.Errorf("missing input")
	}

	if err := C.loadkey(); err != nil {
		return err
	}

	// select action
	var cmd Command
	switch {
	case C.Src != "":
		c := NewEdits()
		if err := c.SetInput(C.Src); err != nil {
			return err
		}
		c.UpdateSrc = true
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
