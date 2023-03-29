package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gregoryv/cmdline"
)

func main() {
	var (
		cli     = cmdline.NewBasicParser()
		keyfile = cli.Option(
			"-a, --api-key-file, $OPENAI_API_KEY_FILE",
		).String(
			os.ExpandEnv("$HOME/.openai.key"),
		)
		model     = "gpt-3.5-turbo"
		role      = "user"
		host      = "api.openai.com"
		inputFile = cli.Option("-i, --input").String("")
	)
	u := cli.Usage()
	u.Example("Ask a question",
		"$ can why is the number 42 significant?",
	)

	u.Example("Provide context",
		"$ correct spelling in ./myfile.txt",
	)
	cli.Parse()

	args := cli.Args()
	if len(args) == 0 {
		log.Fatal("missing input; use --help for usage information")
	}

	// load api key
	key, err := os.ReadFile(keyfile)
	if err != nil {
		log.Fatal(err)
	}

	// decide if we should use edits or chat API

	content := strings.Join(cli.Args(), " ")
	if inputFile != "" {
		v, err := os.ReadFile(inputFile)
		if err != nil {
			log.Fatal(err)
		}
		input := Input{
			"model":       "text-davinci-edit-001",
			"input":       string(v),
			"instruction": content,
		}
		// as json
		data, err := json.Marshal(input)
		if err != nil {
			log.Fatal(err)
		}
		// create api request
		r, _ := http.NewRequest("POST", "/v1/edits", bytes.NewReader(data))
		r.Header.Set("content-type", "application/json")
		r.Header.Set("authorization", "Bearer "+string(key))
		r.URL.Scheme = "https"
		r.URL.Host = host

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

		fmt.Println(result.Choices[0].Text)

	} else {
		// /v1/chat/completions

		// create input
		input := Input{
			"model": model,
			"messages": []Input{
				{
					"role":    role,
					"content": content,
				},
			},
		}
		// as json
		data, err := json.Marshal(input)
		if err != nil {
			log.Fatal(err)
		}

		// create api request
		r, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(data))
		r.Header.Set("content-type", "application/json")
		r.Header.Set("authorization", "Bearer "+string(key))
		r.URL.Scheme = "https"
		r.URL.Host = host

		// send request
		resp, err := http.DefaultClient.Do(r)
		if err != nil {
			log.Fatal(err)
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
		json.NewDecoder(resp.Body).Decode(&result)

		// assuming there will always be at least one choice
		fmt.Println(result.Choices[0].Message.Content)
	}
}

type Input map[string]any
