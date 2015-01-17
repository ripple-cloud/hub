package upstream

import "github.com/ripple-cloud/common/message"

type MockUpstream struct {
	err error
}

func NewMockUpstream() *MockUpstream {
	return &MockUpstream{}
}

func (up *MockUpstream) SetError(e error) {
	up.err = e
}

func (up *MockUpstream) ClearError() {
	up.err = nil
}

func (up *MockUpstream) Connect(address, id string) error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *MockUpstream) Register(msg *message.Message) error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *MockUpstream) Publish(msg *message.Message) error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *MockUpstream) Deregister(msg *message.Message) error {
	if up.err != nil {
		return up.err
	}
	return nil
}

func (up *MockUpstream) Disconnect() {
}
