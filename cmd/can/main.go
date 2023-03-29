package main

import (
	"github.com/gregoryv/cmdline"
)

func main() {
	var (
		cli = cmdline.NewBasicParser()
	)
	cli.Parse()
}
