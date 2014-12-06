package main

import "github.com/ripple/message"

type mqttUpstream struct{}

func newMQTTUpstream() *mqttUpstream {
	return &mqttUpstream{}
}

func (up *mqttUpstream) Connect() error {
	return nil
}

func (up *mqttUpstream) Register(msg *message.Message) error {
	return nil
}

func (up *mqttUpstream) Publish(msg *message.Message) error {
	return nil
}

func (up *mqttUpstream) Deregister(msg *message.Message) error {
	return nil
}
