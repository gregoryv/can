package can

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

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
