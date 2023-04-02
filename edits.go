package can

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func newEdits() *edits {
	return &edits{
		Model:       "text-davinci-edit-001",
		Instruction: "echo",
	}
}

type edits struct {
	Model       string
	input       string
	Instruction string

	// update input file
	UpdateSrc bool

	// result destination
	Out io.Writer

	// path to file
	src       string
	srcIsFile bool
}

// SetInput sets the input to v. If v is a file the content is of that
// file is used.
func (c *edits) SetInput(v string) error {
	if isFile(v) {
		c.src = v
		c.srcIsFile = true
		data, err := os.ReadFile(v)
		if err != nil {
			return fmt.Errorf("SetInput %w", err)
		}
		c.input = string(data)
	} else {
		c.input = v
	}
	return nil
}

func (c *edits) MakeRequest() *http.Request {
	input := map[string]any{
		"model":       c.Model,
		"input":       c.input,
		"instruction": c.Instruction,
	}
	data := should(json.Marshal(input))
	body := bytes.NewReader(data)
	r, _ := http.NewRequest("POST", "/v1/edits", body)
	r.Header.Set("content-type", "application/json")
	return r
}

func (c *edits) HandleResponse(body io.Reader) error {
	// parse result
	var result struct {
		Choices []struct{ Text string }
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("Edits.HandleResponse: %w", err)
		}
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("Edits.HandleResponse: no choices")
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
