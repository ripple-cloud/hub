package upstream

// TODO: Set QOS and add a store for messages

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ripple-cloud/common/message"
)

type listener struct {
	network string
	address string
}

type MQTTUpstream struct {
	id        string
	client    *mqtt.MqttClient
	listeners map[string][]listener
}

func NewMQTTUpstream() *MQTTUpstream {
	return &MQTTUpstream{
		listeners: map[string][]listener{},
	}
}

func (up *MQTTUpstream) Connect(address, id string) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(address)
	opts.SetClientId(id)

	cl := mqtt.NewClient(opts)
	// TODO handle in-flight messages
	_, err := cl.Start()
	if err != nil {
		return err
	}

	up.id = id
	up.client = cl
	return nil
}

// TODO: Change this to a persistent backend
func (up *MQTTUpstream) addListener(topic, network, address string) error {
	up.listeners[topic] = append(up.listeners[topic], listener{network, address})
	return nil
}

func (up *MQTTUpstream) messageHandler(client *mqtt.MqttClient, msg mqtt.Message) {
	// decode the message from payload
	req, err := message.Decode(bytes.NewReader(msg.Payload()))
	if err != nil {
		log.Printf("[error] upstream: failed to decode the message %v err: %s", req, err)
	}

	// ignore other message types
	if req.Type != message.Request {
		log.Printf("[error] upstream: received a message with invalid type %v", req)
	}

	topic, ok := req.Meta["topic"]
	if !ok {
		log.Printf("[error] upstream: missing topic in message %v", req)
	}

	ls := up.listeners[topic]
	for _, l := range ls {
		go func(cl listener) {
			fmt.Println("sending message to listener ", cl)
			// dial the listener
			c, err := net.Dial(cl.network, cl.address)
			if err != nil {
				log.Printf("[error] upstream: failed to dial listener %s err: %s", topic, err)
			}
			defer c.Close()

			// write the message
			// need to re-encode cos handler might add extra meta data
			b, err := req.Encode()
			if err != nil {
				log.Printf("[error] upstream: failed to encode message %v err: %s", req, err)
			}
			if _, err = c.Write(b); err != nil {
				log.Printf("[error] upstream: failed to write the message to %s err: %s", topic, err)
			}
		}(l)
	}
}

// Register for a topic
func (up *MQTTUpstream) Register(msg *message.Message) error {
	topic := msg.Meta["topic"]
	if topic == "" {
		return RequiredFieldMissingError{"topic"}
	}

	// store service's listening interface
	if msg.Meta["network"] != "" && msg.Meta["address"] != "" {
		if err := up.addListener(topic, msg.Meta["network"], msg.Meta["address"]); err != nil {
			return err
		}
	}

	// a service should subscribe to following topics:
	// :topic
	// hub/:hid/:topic
	// TODO: QOS should be set to 1
	tf1, err := mqtt.NewTopicFilter(fmt.Sprintf("hub/%s/%s", up.id, topic), byte(mqtt.QOS_ZERO))
	if err != nil {
		return err
	}

	tf2, err := mqtt.NewTopicFilter(topic, byte(mqtt.QOS_ZERO))
	if err != nil {
		return err
	}

	rcpt, err := up.client.StartSubscription(up.messageHandler, tf1, tf2)
	if err != nil {
		return err
	}

	// return only after receiving the receipt
	<-rcpt
	return nil
}

func (up *MQTTUpstream) Deregister(msg *message.Message) error {
	topic := msg.Meta["topic"]
	if topic == "" {
		return RequiredFieldMissingError{"topic"}
	}

	// remove service from listeners
	ls := up.listeners[topic]
	for i, l := range ls {
		if l.network == msg.Meta["network"] && l.address == msg.Meta["address"] {
			// delete the matching listener
			up.listeners[topic] = append(ls[:i], ls[i+1:]...)
		}
	}
	return nil
}

func (up *MQTTUpstream) Publish(msg *message.Message) error {
	topic := msg.Meta["topic"]
	if topic == "" {
		return RequiredFieldMissingError{"topic"}
	}

	b, err := msg.Encode()
	if err != nil {
		return err
	}
	up.client.PublishMessage(fmt.Sprintf("data/hub/%s/%s", up.id, topic), mqtt.NewMessage(b))
	return nil
}

func (up *MQTTUpstream) Disconnect() {
	// TODO: allow configuring the wait time for disconnect
	up.client.Disconnect(5)
}
