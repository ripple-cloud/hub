package downstream

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/ripple-cloud/common/message"
	"github.com/ripple-cloud/hub/upstream"
)

var up *upstream.MockUpstream = upstream.NewMockUpstream()
var network = "tcp4"
var laddr = ":60001"

func TestListen(t *testing.T) {
	var err chan error
	go func() {
		err <- Listen(network, laddr, up)
	}()
	select {
	case <-err:
		t.Fatalf("failed to listen to downstream requests on %s due to %s", laddr, err)
	case <-time.After(50 * time.Millisecond):
		return
	}
}

func doRequestRaw(req []byte) (*message.Message, error) {
	// dial tcp address
	c, err := net.Dial(network, laddr)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if _, err = c.Write(req); err != nil {
		return nil, err
	}

	// read the response
	b, err := bufio.NewReader(c).ReadBytes('\n')
	if b == nil && err != nil {
		return nil, err
	}

	return message.Decode(bytes.NewReader(b))
}

func doRequest(req *message.Message) (*message.Message, error) {
	// send message
	b, err := req.Encode()
	if err != nil {
		return nil, err
	}

	return doRequestRaw(b)
}

func TestInvalidJSONMessage(t *testing.T) {
	res, err := doRequestRaw([]byte("meta:{},body:test"))
	if err != nil {
		t.Fatal(err)
	}

	if res.Type != message.Error || res.Meta["error"] != "invalid_request" {
		t.Error("expected to receive an error message from server, but got %v", res)
	}
}

func TestRegisterSuccess(t *testing.T) {
	res, err := doRequest(message.NewRegister("app_001", map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Ack {
		t.Errorf("expected to receive an ack message from server, but got %s", res.Type)
	}
}

func TestRegisterError(t *testing.T) {
	up.SetError(errors.New("register_failed"))
	defer up.ClearError()

	res, err := doRequest(message.NewRegister("app_001", map[string]string{}))
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Error || res.Meta["error"] != "register_failed" {
		t.Error("expected to receive an error message from server, but got %v", res)
	}
}

func TestDeregisterSuccess(t *testing.T) {
	res, err := doRequest(message.NewDeregister("app_001"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Ack {
		t.Errorf("expected to receive an ack message from server, but got %s", res.Type)
	}
}

func TestDeregisterError(t *testing.T) {
	up.SetError(errors.New("deregister_failed"))
	defer up.ClearError()

	res, err := doRequest(message.NewDeregister("app_001"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Error || res.Meta["error"] != "deregister_failed" {
		t.Error("expected to receive an error message from server, but got %v", res)
	}
}

func TestPublishSuccess(t *testing.T) {
	req := message.NewPublish("app_001", map[string]string{}, []byte("hello"))

	res, err := doRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Ack {
		t.Errorf("expected to receive an ack message from server, but got %s", res.Type)
	}
}

func TestPublishError(t *testing.T) {
	up.SetError(errors.New("publish_failed"))
	defer up.ClearError()

	req := message.NewPublish("app_001", map[string]string{}, []byte("hello"))

	res, err := doRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Error || res.Meta["error"] != "publish_failed" {
		t.Error("expected to receive an error message from server, but got %v", res)
	}
}

func TestUnknownMessage(t *testing.T) {
	req := message.New()
	req.Type = 9999 //unknown message type
	req.Meta["id"] = "app_001"
	req.Body = []byte("hello")

	res, err := doRequest(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Type != message.Error || res.Meta["error"] != "unknown_message_type" {
		t.Error("expected to receive an error message from server, but got %v", res)
	}
}
