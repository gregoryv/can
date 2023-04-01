// Command can provides access to openai API's from the terminal.
package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/gregoryv/cmdline"
)

var debugOn bool

func main() {
	var (
		cli     = cmdline.NewBasicParser()
		keyfile = cli.Option(
			"-a, --api-key-file, $OPENAI_API_KEY_FILE",
		).String(
			os.ExpandEnv("$HOME/.openai.key"),
		)
		src     = cli.Option("-in", "path to file or block of text").String("")
		debugOn = cli.Flag("--debug")
	)
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

	switch {
	case src != "":
		c := NewEdits()
		c.Src = src
		c.APIKey = string(key)
		c.UpdateSrc = true
		c.Instruction = strings.Join(cli.Args(), " ")
		if err := c.Run(); err != nil {
			log.Fatal(err)
		}
	default:
		c := NewChat()
		c.Content = strings.Join(cli.Args(), " ")
		c.APIKey = string(key)
		if err := c.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

var debug = log.New(ioutil.Discard, "can debug ", log.LstdFlags)
