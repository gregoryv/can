package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func NewChat() *Chat {
	return &Chat{
		API:     "https://api.openai.com/v1/chat/completions",
		Model:   "gpt-3.5-turbo",
		Content: "say hello world!",
		Out:     os.Stdout,
	}
}

type Chat struct {
	API     string // e.g. https://api.openapi.com/v1/chat/completions
	APIKey  string
	Model   string
	Content string
	Out     io.Writer
}

func (c *Chat) Run() error {
	// create input
	input := map[string]any{
		"model": c.Model,
		"messages": []map[string]any{
			{
				"role":    "user",
				"content": c.Content,
			},
		},
	}
	// as json
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}

	// create api request
	r, _ := http.NewRequest("POST", c.API, bytes.NewReader(data))
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer "+c.APIKey)

	// send request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// parse result
	var result struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// assuming there will always be at least one choice
	_, err = c.Out.Write([]byte(result.Choices[0].Message.Content))
	return err
}
