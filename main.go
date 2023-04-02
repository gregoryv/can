// Command can provides access to openai API's from the terminal.
package main

import (
	"log"
	"os"
	"strings"

	"github.com/gregoryv/cmdline"
)

func main() {
	cli := cmdline.NewBasicParser()

	// Skippy the magnificent
	var c Can

	c.SysContent = cli.Option("--system-content, $CAN_SYSTEM_CONTENT").String("")
	c.Src = cli.Option("-in", "path to file or block of text").String("")
	c.API.KeyFile = cli.Option(
		"--api-key-file, $OPENAI_API_KEY_FILE",
	).String(
		os.ExpandEnv("$HOME/.openai.key"),
	)
	c.API.Key = cli.Option("--api-key, $OPENAI_API_KEY").String("")
	debugOn = cli.Flag("--debug")

	u := cli.Usage()
	u.Example("Ask a question",
		"$ can why is the number 42 significant?",
	)
	u.Example("Provide context",
		"$ can correct spelling -in ./README.md",
		"$ can correct spelling -in \"hallo warld\"",
		`$ CAN_SYSTEM_CONTENT="You are a helpful assistant" can Who won the world series in 2020?`,
	)
	cli.Parse()

	c.Input = strings.Join(cli.Args(), " ")

	log.SetFlags(0)

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
