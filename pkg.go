package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
)


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
