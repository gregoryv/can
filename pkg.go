package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
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

func readClose(in io.ReadCloser) *bytes.Buffer {
	var buf bytes.Buffer
	io.Copy(&buf, in)
	in.Close()

	if debugOn {
		var tidy bytes.Buffer
		json.Indent(&tidy, buf.Bytes(), "", "  ")
		debug.Print(tidy.String())
	}
	return &buf
}
