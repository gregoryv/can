package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func NewChat() *Chat {
	return &Chat{
		Model:   "gpt-3.5-turbo",
		Content: "say hello world!",
		Out:     os.Stdout,
	}
}

type Chat struct {
	Model   string
	Content string

	// result destination
	Out io.Writer
}

func (c *Chat) makeRequest() (*http.Request, error) {
	input := map[string]any{
		"model": c.Model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": c.Content,
			},
		},
	}
	data, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("makeRequest %w", err)
	}
	body := bytes.NewReader(data)
	r, _ := http.NewRequest(
		"POST", "https://api.openai.com/v1/chat/completions", body,
	)
	r.Header.Set("content-type", "application/json")
	return r, nil
}

func (c *Chat) handleResponse(body io.Reader) error {
	// parse result
	var result struct {
		Choices []struct{ Message struct{ Content string } }
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return err
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("no choices")
	}

	// act on result
	_, err := c.Out.Write([]byte(result.Choices[0].Message.Content))
	return err
}

// ----------------------------------------

func readClose(in io.ReadCloser) *bytes.Buffer {
	var buf bytes.Buffer
	io.Copy(&buf, in)
	in.Close()

	if debugOn {
		var tidy bytes.Buffer
		json.Indent(&tidy, buf.Bytes(), "", "  ")
		debug.Print(tidy.String())
	}
	return &buf
}
