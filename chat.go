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
	API    string // e.g. https://api.openapi.com/v1/chat/completions
	APIKey string

	Model   string
	Content string

	// result destination
	Out io.Writer
}

func (c *Chat) Run() error {
	r, err := c.makeRequest()
	if err != nil {
		return err
	}

	body, err := sendRequest(r)
	if err != nil {
		return err
	}

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
	_, err = c.Out.Write([]byte(result.Choices[0].Message.Content))
	return err
}

func (c *Chat) makeRequest() (*http.Request, error) {
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
		return nil, fmt.Errorf("makeRequest %w", err)
	}

	// create api request
	r, _ := http.NewRequest("POST", c.API, bytes.NewReader(data))
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer "+c.APIKey)
	return r, nil
}

func sendRequest(r *http.Request) (body *bytes.Buffer, err error) {
	// send request
	debug.Println(r.Method, r.URL)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("sendRequest %w", err)
	}
	debug.Print(resp.Status)

	body = readClose(resp.Body)
	if resp.StatusCode >= 400 {
		log.Print(body.String())
		return nil, fmt.Errorf(resp.Status)
	}
	return
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
