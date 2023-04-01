// Command can provides access to openai API's from the terminal.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
		if err := c.SetSrc(src); err != nil {
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

var (
	debugOn bool
	debug   = log.New(ioutil.Discard, "can debug ", log.LstdFlags)
)

type Command interface {
	MakeRequest() *http.Request
	HandleResponse(io.Reader) error
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
