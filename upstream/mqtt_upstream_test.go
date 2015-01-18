package upstream

// TODO: refactor these tests to be self-contained

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ripple-cloud/common/message"
)

var tester *mqtt.MqttClient
var up *MQTTUpstream

const hubID = "hub-1"
const brokerAddr = "tcp://128.199.132.229:60000" // FIXME: get this from an env variable

func init() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerAddr)
	opts.SetClientId("test-upstream-client")
	opts.SetCleanSession(true)

	tester = mqtt.NewClient(opts)
	_, err := tester.Start()
	if err != nil {
		panic(err)
	}
}

func TestConnect(t *testing.T) {
	up = NewMQTTUpstream()
	if err := up.Connect(brokerAddr, hubID); err != nil {
		t.Fatalf("failed to connect to MQTT server err: %v", err)
	}
}

func startListen(l net.Listener, done chan struct{}, errChan chan error) {
	receivedCount := 0
	fmt.Println("client listening for requests on ", l.Addr())

	for {
		conn, err := l.Accept()
		if err != nil {
			errChan <- err
			return
		}

		// handle request
		go func() {
			defer conn.Close()

			req, err := message.Decode(conn)
			if err != nil {
				errChan <- err
				return
			}

			if req.Type == message.Request {
				receivedCount = receivedCount + 1

				// notify after receiving 2 msgs
				if receivedCount == 2 {
					// stop listening for more requests
					l.Close()

					done <- struct{}{}
				}
			} else {
				errChan <- fmt.Errorf("unexpected message type. expected %s got %s", message.Request, req.Type)
			}
		}()
	}
}

func TestRegister(t *testing.T) {
	topic := "topic-for-register"

	// start a service listening on any available port
	l, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Fatalf("cannot listen on given address err: %s", err)
		return
	}
	defer l.Close()

	done := make(chan struct{})
	errChan := make(chan error)
	go startListen(l, done, errChan)

	// register the service with upstream
	regMsg := message.NewRegister(topic, map[string]string{
		"network": l.Addr().Network(),
		"address": l.Addr().String(),
	})
	if err := up.Register(regMsg); err != nil {
		t.Fatalf("register failed - %v", err)
	}

	// send requests on a topic
	reqMsg := message.NewRequest(topic, map[string]string{}, []byte("hello"))
	b, err := reqMsg.Encode()
	if err != nil {
		t.Fatalf("failed to encode message %s", err)
	}
	m := mqtt.NewMessage(b)
	tester.PublishMessage(fmt.Sprintf("hub/%s/%s", hubID, topic), m)
	tester.PublishMessage(fmt.Sprintf("%s", topic), m)

	select {
	case <-done:
		// service received all messages
		// test passes
		return
	case err := <-errChan:
		t.Error(err)
	case <-time.After(10 * time.Second):
		t.Error("timed out waiting for messages")
	}
}

func TestRegisterWithoutTopic(t *testing.T) {
	msg := message.New()
	msg.Type = message.Register

	err := up.Register(msg)
	if _, ok := err.(RequiredFieldMissingError); !ok {
		t.Errorf("expected register to return required field missing error, but got - %v", err)
	}
}

func TestDeregister(t *testing.T) {
	topic := "topic-for-deregister"

	// register 3 services for same topic
	for _, s := range [][]string{{"tcp", ":8080"}, {"tcp", "9000"}, {"tcp", "11000"}} {
		regMsg := message.NewRegister(topic, map[string]string{
			"network": s[0],
			"address": s[1],
		})
		if err := up.Register(regMsg); err != nil {
			t.Fatalf("register failed - %v", err)
		}
	}

	if len(up.listeners[topic]) != 3 {
		t.Error("not all services registered")
	}

	// deregister one service
	deregMsg := message.NewDeregister(topic, map[string]string{
		"network": "tcp",
		"address": ":9000",
	})
	if err := up.Deregister(deregMsg); err != nil {
		t.Fatalf("deregister failed - %v", err)
	}

	// should only remove that service's listener
	ls := up.listeners[topic]
	if ls[0].network != "tcp" || ls[1].address != ":8080" &&
		ls[1].network != "tcp" && ls[1].address != ":11000" {
		t.Error("should not have remove all listeners for the topic")
	}
}

func TestDeregisterWithoutTopic(t *testing.T) {
	msg := message.New()
	msg.Type = message.Deregister

	err := up.Deregister(msg)
	if _, ok := err.(RequiredFieldMissingError); !ok {
		t.Errorf("expected deregister to return required field missing error, but got - %v", err)
	}
}

func TestPublish(t *testing.T) {
	topic := "topic-for-publish"

	// subscribe to messages for the topic
	tf, err := mqtt.NewTopicFilter(fmt.Sprintf("data/hub/%s/%s", up.id, topic), byte(mqtt.QOS_ZERO))
	if err != nil {
		t.Fatalf("failed to create topic filter %s", err)
	}

	rcvdMsg := make(chan *message.Message)
	errChan := make(chan error)
	handler := func(client *mqtt.MqttClient, msg mqtt.Message) {
		pm, err := message.Decode(bytes.NewReader(msg.Payload()))
		if err != nil {
			errChan <- err
		}

		rcvdMsg <- pm
	}
	rcpt, err := tester.StartSubscription(handler, tf)
	if err != nil {
		t.Errorf("failed to subscribe to topic %s", err)
	}

	// wait for the receipt
	<-rcpt

	// publish a message to upstream
	pubMsg := message.NewPublish(topic, map[string]string{"foo": "bar"}, []byte("hello"))
	if err := up.Publish(pubMsg); err != nil {
		t.Fatalf("publish failed - %v", err)
	}

	select {
	case m := <-rcvdMsg:
		if m.Type != message.Publish {
			t.Errorf("expected a message with type publish: %v", m)
		}
		if !reflect.DeepEqual(m.Meta, map[string]string{"foo": "bar", "topic": topic}) {
			t.Errorf("received message has unexpected meta data: %v", m)
		}
		if !bytes.Equal(m.Body, []byte("hello")) {
			t.Errorf("received message has an unexpected body %v", m)
		}
	case err := <-errChan:
		t.Error(err)
	case <-time.After(10 * time.Second):
		t.Error("timed out waiting for messages")
	}
}

func TestPublishWithoutTopic(t *testing.T) {
	msg := message.New()
	msg.Type = message.Publish

	err := up.Publish(msg)
	if _, ok := err.(RequiredFieldMissingError); !ok {
		t.Errorf("expected publish to return required field missing error, but got - %v", err)
	}
}

func TestDisconnect(t *testing.T) {
	up.Disconnect()

	// TODO: verify if disconnected from upstream
}

func TestTeardown(t *testing.T) {
	tester.Disconnect(5)
}
