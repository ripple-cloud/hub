package main

import (
	"testing"
	"time"
)

var up Upstream = newMockUpstream()
var network = "tcp4"
var laddr = ":60001"

func TestListenDownstream(t *testing.T) {
	var err chan error
	go func() {
		err <- listenDownstream(network, laddr, up)
	}()
	select {
	case <-err:
		t.Fatalf("failed to listen to downstream requests on %s due to %s", laddr, err)
	case <-time.After(50 * time.Millisecond):
		return
	}
}

func TestInvalidJSONMessage(t *testing.T) {
	// dial tcp address
	// send message
	// check the response
}

func TestRegisterMessage(t *testing.T) {
	// test for success
	// test for failure
}

func TestDeregisterMessage(t *testing.T) {

}

func TestPublishMessage(t *testing.T) {

}

func TestUnknownMessage(t *testing.T) {
}
