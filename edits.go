package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func NewEdits() *Edits {
	return &Edits{
		Model:       "text-davinci-edit-001",
		Instruction: "echo",
	}
}

type Edits struct {
	Model string
	// if Src is a file should the output be written to the same file
	UpdateSrc bool

	Instruction string

	// result destination
	Out io.Writer

	// path to file or block of text
	src       string
	srcIsFile bool
}

func (c *Edits) SetSrc(v string) error {
	if isFile(v) {
		c.srcIsFile = true
		data, err := os.ReadFile(c.src)
		if err != nil {
			return fmt.Errorf("SetSrc %w", err)
		}
		c.src = string(data)
	} else {
		c.src = v
	}
	return nil
}

func (c *Edits) MakeRequest() *http.Request {
	input := map[string]any{
		"model":       c.Model,
		"input":       c.src,
		"instruction": c.Instruction,
	}
	data := should(json.Marshal(input))
	body := bytes.NewReader(data)
	r, _ := http.NewRequest(
		"POST", "https://api.openai.com/v1/edits", body,
	)
	r.Header.Set("content-type", "application/json")
	return r
}

func (c *Edits) HandleResponse(body io.Reader) error {
	// parse result
	var result struct {
		Choices []struct{ Text string }
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return err
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("no choices")
	}

	// act on result
	if c.srcIsFile && c.UpdateSrc {
		out, err := os.Create(c.src)
		if err != nil {
			return err
		}
		c.Out = out
	}
	if c.Out == nil {
		c.Out = os.Stdout
	}
	_, err := c.Out.Write([]byte(result.Choices[0].Text))
	return err
}
