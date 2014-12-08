package upstream

import "github.com/ripple/message"

type MQTTUpstream struct{}

func NewMQTTUpstream() *MQTTUpstream {
	return &MQTTUpstream{}
}

func (up *MQTTUpstream) Connect() error {
	return nil
}

func (up *MQTTUpstream) Register(msg *message.Message) error {
	return nil
}

func (up *MQTTUpstream) Publish(msg *message.Message) error {
	return nil
}

func (up *MQTTUpstream) Deregister(msg *message.Message) error {
	return nil
}
