// Command can provides access to openai API's from the terminal.
package main

import (
	"log"
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
		inputFile = cli.Option("-i, --input").String("")
		update    = cli.Option("-u, --update", "write result to input file").Bool()
	)
	u := cli.Usage()
	u.Example("Ask a question",
		"$ can why is the number 42 significant?",
	)

	u.Example("Provide context",
		"$ can correct spelling in -i ./README.md -u",
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

	switch {
	case inputFile != "":
		c := NewEdits()
		c.InputFile = inputFile
		c.APIKey = string(key)
		c.Update = update
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
