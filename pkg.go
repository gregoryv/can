package can

import (
	"io/ioutil"
	"log"
	"os"
)

func SetDebug(v bool) {
	if v {
		debugOn = true
		debug.SetOutput(os.Stderr)
		return
	}
	debugOn = false
	debug.SetOutput(ioutil.Discard)
}

var (
	debugOn bool
	debug   = log.New(ioutil.Discard, "can debug ", log.LstdFlags)
)

func isFile(src string) bool {
	_, err := os.Stat(src)
	return err == nil
}

func should(data []byte, err error) []byte {
	if err != nil {
		log.Print(err)
	}
	return data
}
