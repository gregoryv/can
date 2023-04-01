package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	args := []string{"can", "--debug", "what is your favourite color?"}
	if testing.Short() {
		args = []string{"can", "-h"}
	}
	defer func() { _ = recover() /* catch expected panic */ }()
	os.Args = args
	main()
}
