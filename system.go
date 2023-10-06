/*
Package can provides a system for interacting with api.openai.com/v1

The main purpose of this system is to support a simple command line
application (cmd/can). It is Not a generic client for api.openai.com
*/
package can

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

// NewSystem returns a system with debug log disabled.
func NewSystem() *System {
	return &System{
		debug: log.New(ioutil.Discard, "can debug ", log.Flags()),
	}
}

type System struct {
	api struct {
		*url.URL // host:port
		KeyFile  string
		Key      string
	}
	sysContent string
	updateSrc  bool
	src        string // ie. file or block of text
	input      string

	debugOn bool
	debug   *log.Logger
}

// system settings

// SetAPIUrl use scheme://hostname[:port], e.g. https://api.openai.com
func (s *System) SetAPIUrl(v *url.URL) { s.api.URL = v }

// SetAPIKey sets the bearer key used in API calls.
func (s *System) SetAPIKey(v string) { s.api.Key = v }

// SetAPIKeyFile sets the path to the file containing the API Key. If
// SetAPIKey is used this has no effect.
func (s *System) SetAPIKeyFile(v string) { s.api.KeyFile = v }

// I dislike these settings; they are unclear

// SetSysContent sets the system role content for /v1/chat/completions.
func (s *System) SetSysContent(v string) { s.sysContent = v }

// SetSrc sets the path or text used for /v1/edits input.
func (s *System) SetSrc(v string) { s.src = v }

// SetUpdateSrc to control if result from /v1/edits should be written
// back to the src file. Has no effect if the src is not a file.
func (s *System) SetUpdateSrc(v bool) { s.updateSrc = v }

// SetInput sets the question or instruction.
func (s *System) SetInput(v string) { s.input = v }

// SetDebugOutput writer, use nil to disable
func (s *System) SetDebugOutput(v io.Writer) {
	if v == nil {
		s.debugOn = false
		s.debug.SetOutput(ioutil.Discard)
		return
	}
	s.debugOn = true
	s.debug.SetOutput(v)
}

// Run builds and executes the request according to the system
// settings. The idea is to call Run once. It is Not safe to call in
// concurrent goroutines.
func (s *System) Run() error {
	if len(s.input) == 0 {
		return fmt.Errorf("missing input")
	}
	if err := s.loadkey(); err != nil {
		return err
	}
	if s.api.URL == nil {
		return fmt.Errorf("Can.Run: missing API.URL")
	}
	if s.debug == nil {
		s.debug = log.New(ioutil.Discard, "can debug ", log.Flags())
	}

	// select action
	var cmd command
	switch {
	case s.src != "":
		c := newEdits()
		if err := c.SetInput(s.src); err != nil {
			return err
		}
		c.UpdateSrc = s.updateSrc
		c.Instruction = s.input
		cmd = c

	default:
		c := newChat()
		c.Content = s.input
		c.SystemContent = s.sysContent
		cmd = c
	}

	// execute action
	r := cmd.MakeRequest()
	r.Header.Set("authorization", "Bearer "+s.api.Key)
	r.URL.Host = s.api.URL.Host
	r.URL.Scheme = s.api.URL.Scheme

	body, err := s.sendRequest(r)
	if err != nil {
		return err
	}

	return cmd.HandleResponse(body)
}

func (s *System) loadkey() error {
	if s.api.Key != "" {
		return nil
	}
	if s.api.KeyFile == "" {
		return nil
	}
	data, err := os.ReadFile(s.api.KeyFile)
	if err != nil {
		return err
	}
	s.api.Key = string(bytes.TrimSpace(data))
	return nil
}

// sendRequest sends the given request and returns the body. HTTP
// status code >= 400 result in errors.
func (s *System) sendRequest(r *http.Request) (body *bytes.Buffer, err error) {
	s.debug.Println(r.Method, r.URL)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("sendRequest %w", err)
	}
	s.debug.Print(resp.Status)

	body = s.readClose(resp.Body)
	if resp.StatusCode >= 400 {
		log.Print(body.String())
		return nil, fmt.Errorf(resp.Status)
	}
	return
}

func (s *System) readClose(in io.ReadCloser) *bytes.Buffer {
	var buf bytes.Buffer
	io.Copy(&buf, in)
	in.Close()

	if s.debugOn {
		var tidy bytes.Buffer
		json.Indent(&tidy, buf.Bytes(), "", "  ")
		s.debug.Print(tidy.String())
	}
	return &buf
}

// ----------------------------------------

// command represents an API call with a parsing method
type command interface {
	MakeRequest() *http.Request
	HandleResponse(io.Reader) error
}

func newChat() *chat {
	return &chat{
		Model:   "gpt-3.5-turbo",
		Content: "say hello world!",
	}
}

type chat struct {
	Model         string
	Content       string
	SystemContent string

	// result destination
	Out io.Writer
}

// MakeRequest returns a request for /v1/chat/completions
func (c *chat) MakeRequest() *http.Request {
	type m struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	messages := []m{m{"user", c.Content}}
	if v := c.SystemContent; v != "" {
		messages = append(messages, m{"system", v})
	}
	input := map[string]any{
		"model":    c.Model,
		"messages": messages,
	}
	data := should(json.Marshal(input))
	body := bytes.NewReader(data)
	r, _ := http.NewRequest("POST", "/v1/chat/completions", body)
	r.Header.Set("content-type", "application/json")
	return r
}

// HandleResponse writes the first choice, returns an error if there
// are no choices.
func (c *chat) HandleResponse(body io.Reader) error {
	// parse result
	var result struct {
		Choices []struct{ Message struct{ Content string } }
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("Chat.HandleResponse: no choices")
	}
	if c.Out == nil {
		c.Out = os.Stdout
	}

	// act on result
	_, err := c.Out.Write([]byte(result.Choices[0].Message.Content))
	return err
}

// ----------------------------------------

func newEdits() *edits {
	return &edits{
		Model:       "text-davinci-edit-001",
		Instruction: "echo",
	}
}

type edits struct {
	// contains mixed public and private settings; maybe convert all to settings

	Model       string
	input       string
	Instruction string

	// update input file
	UpdateSrc bool

	// result destination
	Out io.Writer

	// path to file
	src       string
	srcIsFile bool
}

// SetInput sets the input to v. If v is a file the content of that
// file is used.
func (c *edits) SetInput(v string) error {
	if isFile(v) {
		c.src = v
		c.srcIsFile = true
		data, err := os.ReadFile(v)
		if err != nil {
			return fmt.Errorf("SetInput %w", err)
		}
		c.input = string(data)
	} else {
		c.input = v
	}
	return nil
}

// MakeRequest returns a request for /v1/edits
func (c *edits) MakeRequest() *http.Request {
	input := map[string]any{
		"model":       c.Model,
		"input":       c.input,
		"instruction": c.Instruction,
	}
	data := should(json.Marshal(input))
	body := bytes.NewReader(data)
	r, _ := http.NewRequest("POST", "/v1/edits", body)
	r.Header.Set("content-type", "application/json")
	return r
}

// HandleResponse writes the first choice, returns an error if there
// are no choices.
func (c *edits) HandleResponse(body io.Reader) error {
	// parse result
	var result struct {
		Choices []struct{ Text string }
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("Edits.HandleResponse: %w", err)
		}
	}
	if len(result.Choices) == 0 {
		return fmt.Errorf("Edits.HandleResponse: no choices")
	}

	// act on result
	if c.srcIsFile && c.UpdateSrc {
		out, err := os.Create(c.src)
		if err != nil {
			return err
		}
		c.Out = out
	}
	if c.Out == nil {
		c.Out = os.Stdout
	}
	_, err := c.Out.Write([]byte(result.Choices[0].Text))
	return err
}

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
