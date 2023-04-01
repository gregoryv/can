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
	// path to file or block of text
	Src string
	// if Src is a file should the output be written to the same file
	UpdateSrc bool

	Instruction string

	// result destination
	Out io.Writer
}

func (c *Edits) MakeRequest() (*http.Request, error) {
	v := c.Src
	if isFile(c.Src) {
		data, err := os.ReadFile(c.Src)
		if err != nil {
			return nil, fmt.Errorf("makeRequest %w", err)
		}
		v = string(data)
	}
	input := map[string]any{
		"model":       c.Model,
		"input":       string(v),
		"instruction": c.Instruction,
	}
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("makeRequest %w", err)
	}
	body := bytes.NewReader(data)
	r, _ := http.NewRequest(
		"POST", "https://api.openai.com/v1/edits", body,
	)
	r.Header.Set("content-type", "application/json")
	return r, nil
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
	if isFile(c.Src) && c.UpdateSrc {
		out, err := os.Create(c.Src)
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

func isFile(src string) bool {
	_, err := os.Stat(src)
	return err == nil
}
