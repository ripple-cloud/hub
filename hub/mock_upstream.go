package main

import "github.com/ripple/message"

type mockUpstream struct {
	err error
}

func newMockUpstream() *mockUpstream {
	return &mockUpstream{}
}

func (up *mockUpstream) SetError(e error) {
	up.err = e
}

func (up *mockUpstream) ClearError() {
	up.err = nil
}

func (up *mockUpstream) Connect() error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *mockUpstream) Register(msg *message.Message) error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *mockUpstream) Publish(msg *message.Message) error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *mockUpstream) Deregister(msg *message.Message) error {
	if up.err != nil {
		return up.err
	}
	return nil
}
