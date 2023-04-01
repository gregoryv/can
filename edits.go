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
	API         string // e.g. https://api.openapi.com/v1/chat/completions
	APIKey      string
	Model       string
	InFile      string
	Instruction string

	Update bool
	Out    io.Writer
}

func (e *Edits) Run() error {
	v, err := os.ReadFile(e.InFile)
	if err != nil {
		log.Fatal(err)
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

	if e.Update {
		if out, err := os.Create(e.InFile); err != nil {
			return err
		} else {
			e.Out = out
		}
	}

	if len(result.Choices) == 0 {
		return fmt.Errorf("no choices")
	}
	_, err = e.Out.Write([]byte(result.Choices[0].Text))
	return err
}
