package upstream

// TODO: Set QOS and add a store for messages

import (
	"bytes"
	"fmt"
	"log"
	"net"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ripple-cloud/hub/message"
)

type appListener struct {
	network string
	address string
}

type MQTTUpstream struct {
	id           string
	client       *mqtt.MqttClient
	appListeners map[string]appListener
}

func NewMQTTUpstream() *MQTTUpstream {
	return &MQTTUpstream{
		appListeners: map[string]appListener{},
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
func (up *MQTTUpstream) addAppListener(id, network, address string) error {
	up.appListeners[id] = appListener{network, address}
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

	appID, ok := req.Meta["id"]
	if !ok {
		log.Printf("[error] upstream: missing id in message %v", req)
	}

	l, ok := up.appListeners[appID]
	if ok {
		// dial the listener
		c, err := net.Dial(l.network, l.address)
		if err != nil {
			log.Printf("[error] upstream: failed to dail app listener %s err: %s", appID, err)
		}
		defer c.Close()

		// write the message
		b, err := req.Encode()
		if err != nil {
			log.Printf("[error] upstream: failed to re-encode message %v err: %s", req, err)
		}
		if _, err = c.Write(b); err != nil {
			log.Printf("[error] upstream: failed to write the message to %s err: %s", appID, err)
		}
	}
}

func (up *MQTTUpstream) Register(msg *message.Message) error {
	if msg.Meta["id"] == "" {
		return RequiredFieldMissingError{"id"}
	}

	// store app's listening interface
	if msg.Meta["network"] != "" && msg.Meta["address"] != "" {
		if err := up.addAppListener(msg.Meta["id"], msg.Meta["network"], msg.Meta["address"]); err != nil {
			return err
		}
	}

	// An app should subscribe to following topics:
	// hub/:hid/app/:id
	// app/:id
	appID := msg.Meta["id"]
	// TODO: QOS should be set to 1 or 2
	tf1, err := mqtt.NewTopicFilter(fmt.Sprintf("hub/%s/app/%s", up.id, appID), byte(mqtt.QOS_ZERO))
	if err != nil {
		return err
	}

	tf2, err := mqtt.NewTopicFilter(fmt.Sprintf("app/%s", appID), byte(mqtt.QOS_ZERO))
	if err != nil {
		return err
	}

	if _, err = up.client.StartSubscription(up.messageHandler, tf1, tf2); err != nil {
		return err
	}

	return nil
}

func (up *MQTTUpstream) Deregister(msg *message.Message) error {
	if msg.Meta["id"] == "" {
		return RequiredFieldMissingError{"id"}
	}

	// Unsubscribe an app from following topics:
	// hub/:hid/app/:id
	// app/:id
	appID := msg.Meta["id"]
	t1 := fmt.Sprintf("hub/%s/app/%s", up.id, appID)
	t2 := fmt.Sprintf("app/%s", up.id, appID)
	if _, err := up.client.EndSubscription(t1, t2); err != nil {
		return err
	}

	// remove app from appListeners
	delete(up.appListeners, appID)
	return nil
}

func (up *MQTTUpstream) Publish(msg *message.Message) error {
	if msg.Meta["id"] == "" {
		return RequiredFieldMissingError{"id"}
	}

	appID := msg.Meta["id"]
	b, err := msg.Encode()
	if err != nil {
		return err
	}
	up.client.PublishMessage(fmt.Sprintf("data/hub/%s/app/%s", up.id, appID), mqtt.NewMessage(b))

	return nil
}

func (up *MQTTUpstream) Disconnect() {
	// TODO: allow configuring wait time for disconnect
	up.client.Disconnect(5)
}
