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

func NewEdits() *Edits {
	return &Edits{
		API:         "https://api.openai.com/v1/edits",
		Model:       "text-davinci-edit-001",
		Instruction: "echo",
		Out:         os.Stdout,
	}
}

type Edits struct {
	// e.g. https://api.openapi.com/v1/chat/completions
	API    string
	APIKey string

	Model string
	// path to file or block of text
	Src string
	// if Src is a file should the output be written to the same file
	UpdateSrc bool

	Instruction string

	// result destination
	Out io.Writer
}

func (e *Edits) Run() error {
	v := e.Src
	if isFile(e.Src) {
		data, err := os.ReadFile(e.Src)
		if err != nil {
			return err
		}
		v = string(data)
	}
	input := map[string]any{
		"model":       e.Model,
		"input":       string(v),
		"instruction": e.Instruction,
	}
	// as json
	data, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	// create api request
	r, _ := http.NewRequest("POST", e.API, bytes.NewReader(data))
	r.Header.Set("content-type", "application/json")
	r.Header.Set("authorization", "Bearer "+e.APIKey)

	// send request
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// parse result
	var result struct {
		Choices []struct {
			Text string
		}
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if isFile(e.Src) && e.UpdateSrc {
		out, err := os.Create(e.Src)
		if err != nil {
			return err
		}
		e.Out = out
	}

	if len(result.Choices) == 0 {
		return fmt.Errorf("no choices")
	}
	_, err = e.Out.Write([]byte(result.Choices[0].Text))
	return err
}

func isFile(src string) bool {
	_, err := os.Stat(src)
	return err == nil
}
