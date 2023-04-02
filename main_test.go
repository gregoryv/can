package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	// any call that fails
	os.Args = []string{"can", "--api-url", "http://localhost:12345"}
	fatal = func(...interface{}) { /*noop*/ }
	main()
}
