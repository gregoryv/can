// Command can provides access to openai API's from the terminal.
package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gregoryv/cmdline"
)

func main() {
	cli := cmdline.NewBasicParser()
	src := cli.Option("-in", "path to file or block of text").String("")
	keyfile := cli.Option(
		"--api-key-file, $OPENAI_API_KEY_FILE",
	).String(
		os.ExpandEnv("$HOME/.openai.key"),
	)
	debugOn = cli.Flag("--debug")

	u := cli.Usage()
	u.Example("Ask a question",
		"$ can why is the number 42 significant?",
	)
	u.Example("Provide context",
		"$ can correct spelling -in ./README.md",
		"$ can correct spelling -in \"hallo warld\"",
	)
	cli.Parse()

	log.SetFlags(0)
	if debugOn {
		debug.SetOutput(os.Stderr)
	}

	args := cli.Args()
	if len(args) == 0 {
		log.Fatal("missing input; use --help for usage information")
	}

	// load api key
	key, err := os.ReadFile(keyfile)
	if err != nil {
		log.Fatal(err)
	}
	key = bytes.TrimSpace(key)

	// select action
	var cmd Command
	switch {
	case src != "":
		c := NewEdits()
		if err := c.SetInput(src); err != nil {
			log.Fatal(err)
		}
		c.UpdateSrc = true
		c.Instruction = strings.Join(cli.Args(), " ")
		cmd = c

	default:
		c := NewChat()
		c.Content = strings.Join(cli.Args(), " ")
		cmd = c
	}

	// execute action
	r := cmd.MakeRequest()
	r.Header.Set("authorization", "Bearer "+string(key))

	body, err := sendRequest(r)
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.HandleResponse(body); err != nil {
		log.Fatal(err)
	}
}

type Command interface {
	MakeRequest() *http.Request
	HandleResponse(io.Reader) error
}
