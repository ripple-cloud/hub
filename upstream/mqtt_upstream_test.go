package upstream

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/ripple/message"
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
	appID := "app-001"

	// start an app listening on any available port
	l, err := net.Listen("tcp4", ":0")
	if err != nil {
		t.Fatalf("cannot listen on given address err: %s", err)
		return
	}
	defer l.Close()

	done := make(chan struct{})
	errChan := make(chan error)
	go startListen(l, done, errChan)

	// register the app with upstream
	regMsg := message.NewRegister(appID, map[string]string{
		"network": l.Addr().Network(),
		"address": l.Addr().String(),
	})
	if err := up.Register(regMsg); err != nil {
		t.Fatalf("register failed - %v", err)
	}

	// send requests to the app
	reqMsg := message.NewRequest(appID, map[string]string{}, []byte("hello"))
	b, err := reqMsg.Encode()
	if err != nil {
		t.Fatalf("failed to encode message %s", err)
	}
	m := mqtt.NewMessage(b)
	tester.PublishMessage(fmt.Sprintf("hub/%s/app/%s", hubID, appID), m)
	tester.PublishMessage(fmt.Sprintf("app/%s", appID), m)

	select {
	case <-done:
		// app received all messages
		// test passes
		return
	case err := <-errChan:
		t.Error(err)
	case <-time.After(10 * time.Second):
		t.Error("timed out waiting for messages")
	}
}

func TestRegisterWithoutID(t *testing.T) {
	msg := message.New()
	msg.Type = message.Register

	err := up.Register(msg)
	if _, ok := err.(RequiredFieldMissingError); !ok {
		t.Errorf("expected register to return id required error, but got - %v", err)
	}
}

func TestDeregister(t *testing.T) {
	appID := "app-001"

	// register the app with upstream
	regMsg := message.NewRegister(appID, map[string]string{
		"network": "tcp",
		"address": ":8080",
	})
	if err := up.Register(regMsg); err != nil {
		t.Fatalf("register failed - %v", err)
	}

	_, ok := up.appListeners[appID]
	if !ok {
		t.Error("app's listener should be registered")
	}

	// *deregister* the app
	deregMsg := message.NewDeregister(appID)
	if err := up.Deregister(deregMsg); err != nil {
		t.Fatalf("deregister failed - %v", err)
	}

	// TODO: currently there's no test for UNSUBSCRIBE on MQTT server
	// let's just assume the message is sent correctly

	// should remove app's listener
	_, ok = up.appListeners[appID]
	if ok {
		t.Error("app's listener should be removed after deregister")
	}
}

func TestDeregisterWithoutID(t *testing.T) {
	msg := message.New()
	msg.Type = message.Deregister

	err := up.Deregister(msg)
	if _, ok := err.(RequiredFieldMissingError); !ok {
		t.Errorf("expected register to return id required error, but got - %v", err)
	}
}

func TestPublish(t *testing.T) {
	appID := "app-001"

	// subscribe to notices from the app
	tf, err := mqtt.NewTopicFilter(fmt.Sprintf("data/hub/%s/app/%s", up.id, appID), byte(mqtt.QOS_ZERO))
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
	if _, err = tester.StartSubscription(handler, tf); err != nil {
		t.Errorf("failed to subscribe to topic %s", err)
	}

	// publish a message to upstream
	pubMsg := message.NewPublish(appID, map[string]string{"foo": "bar"}, []byte("hello"))
	if err := up.Publish(pubMsg); err != nil {
		t.Fatalf("publish failed - %v", err)
	}

	select {
	case m := <-rcvdMsg:
		if m.Type != message.Publish {
			t.Errorf("expected a message with type publish: %v", m)
		}
		if !reflect.DeepEqual(m.Meta, map[string]string{"foo": "bar", "id": appID}) {
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

func TestPublishWithoutID(t *testing.T) {
	msg := message.New()
	msg.Type = message.Publish

	err := up.Publish(msg)
	if _, ok := err.(RequiredFieldMissingError); !ok {
		t.Errorf("expected register to return id required error, but got - %v", err)
	}
}

func TestDisconnect(t *testing.T) {
	up.Disconnect()

	// TODO: verify if disconnected from upstream
}

func TestTeardown(t *testing.T) {
	tester.Disconnect(5)
}
