package message

import (
	"encoding/json"
	"io"
)

type Message struct {
	Type int64             `json:"type"`
	Meta map[string]string `json:"meta",omitempty`
	Body []byte            `json:"body",omitempty`
}

const (
	Register = iota + 1
	Deregister
	Publish
)

func Decode(r io.Reader) (*Message, error) {
	m := &Message{}
	err := json.NewDecoder(r).Decode(m)
	return m, err
}