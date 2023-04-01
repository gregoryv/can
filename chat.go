package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	debug.Println(r.Method, r.URL, len(data), "bytes")
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		debug.Print(resp.Status)
		return err
	}

	body := readClose(resp.Body)
	if resp.StatusCode >= 400 {
		log.Print(body.String())
		return fmt.Errorf(resp.Status)
	}

	// parse result
	var result struct {
		Choices []struct {
			Message struct {
				Content string
			}
		}
	}

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return err
	}

	// assuming there will always be at least one choice
	if len(result.Choices) == 0 {
		return fmt.Errorf("no choices")
	}
	_, err = c.Out.Write([]byte(result.Choices[0].Message.Content))
	return err
}

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
