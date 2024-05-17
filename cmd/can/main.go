// Command can provides access to openai API's from the terminal.
package main

import (
	"log"
	"os"
	"strings"

	"github.com/gregoryv/can"
	"github.com/gregoryv/cmdline"
)

func main() {
	cli := cmdline.NewBasicParser()

	// Skippy the magnificent
	s := can.NewSystem()

	var (
		sysContent = cli.Option("--system-content, $CAN_SYSTEM_CONTENT").String("")
		src        = cli.Option("-in",
			"path to file or block of text",
			"result is written on stdout",
		).String("")
		keyFile = cli.Option(
			"--api-key-file, $OPENAI_API_KEY_FILE",
		).String(
			os.ExpandEnv("$HOME/.openai.key"),
		)
		key    = cli.Option("--api-key, $OPENAI_API_KEY").String("")
		apiUrl = cli.Option("--api-url, $OPENAI_API_URL").Url("https://api.openai.com")
	)

	if cli.Flag("--debug") {
		s.SetDebugOutput(os.Stderr)
	}

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

	s.SetAPIKeyFile(keyFile)
	s.SetAPIKey(key)
	s.SetAPIUrl(apiUrl)
	s.SetSysContent(sysContent)
	s.SetSrc(src)
	s.SetUpdateSrc(src != "")

	s.SetInput(strings.Join(cli.Args(), " "))

	log.SetFlags(0)

	if err := s.Run(); err != nil {
		fatal(err)
	}
}

// here so we can fully test func main
var fatal func(...interface{}) = log.Fatal
