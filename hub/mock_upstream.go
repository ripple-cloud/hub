package main

import "github.com/ripple/message"

type mockUpstream struct{}

func newMockUpstream() *mockUpstream {
	return &mockUpstream{}
}

func (up *mockUpstream) Connect() error {
	return nil
}

func (up *mockUpstream) Register(msg *message.Message) error {
	return nil
}

func (up *mockUpstream) Publish(msg *message.Message) error {
	return nil
}

func (up *mockUpstream) Deregister(msg *message.Message) error {
	return nil
}
