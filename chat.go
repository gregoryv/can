package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

func NewChat() *Chat {
	return &Chat{
		Model:   "gpt-3.5-turbo",
		Content: "say hello world!",
	}
}

type Chat struct {
	Model         string
	Content       string
	SystemContent string

	// result destination
	Out io.Writer
}

func (c *Chat) MakeRequest() *http.Request {
	messages := []map[string]any{
		{
			"role":    "user",
			"content": c.Content,
		},
	}
	if v := c.SystemContent; v != "" {
		messages = append(messages, map[string]any{
			"role":    "system",
			"content": v,
		})
	}
	input := map[string]any{
		"model":    c.Model,
		"messages": messages,
	}
	data := should(json.Marshal(input))
	body := bytes.NewReader(data)
	r, _ := http.NewRequest("POST", "/v1/chat/completions", body)
	r.Header.Set("content-type", "application/json")
	return r
}

func (c *Chat) HandleResponse(body io.Reader) error {
	// parse result
	var result struct {
		Choices []struct{ Message struct{ Content string } }
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("Chat.HandleResponse: no choices")
	}
	if c.Out == nil {
		c.Out = os.Stdout
	}

	// act on result
	_, err := c.Out.Write([]byte(result.Choices[0].Message.Content))
	return err
}
