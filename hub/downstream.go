package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/ripple/message"
)

func listenDownstream(network, laddr string, up Upstream) error {
	l, err := net.Listen(network, laddr)
	if err != nil {
		log.Fatalf("[error] %s", err)
	}
	defer l.Close()

	log.Printf("[info] %s is listening on %v", appName, l.Addr())

	for {
		// accept requests and process them
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("[error] %s", err)
		}

		go downstreamHandler(conn, up)
	}
}

// downstreamHandler will relay requests from downstream to upstream
func downstreamHandler(c net.Conn, up Upstream) {
	defer c.Close()

	// decode the incoming message as
	msg, err := message.Decode(c)
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
		return
	}

	switch msg.Type {
	case message.Register:
		// registers with upstream
		// app has to specify the interface its listening in `meta`
		// send an error response if no listening interface provided
		err := up.Register(msg)
		if err != nil {
			respondError(c, err)
		}

		// on success, write { status: "ok } msg
		respondOK(c)
	case message.Publish:
		err := up.Publish(msg)
		if err != nil {
			respondError(c, err)
		}

		// on success, write { status: "ok } msg
		respondOK(c)
	case message.Deregister:
		err := up.Deregister(msg)
		if err != nil {
			respondError(c, err)
		}

		// on success, write { status: "ok } msg
		respondOK(c)
	default:
		log.Printf("[error] [client %s]: received a message with unsupported type - %s", msg.Type)
		return
	}
}

func respondError(c net.Conn, err error) {
	errResponse := map[string]string{
		"error": err.Error(),
	}
	j, err := json.Marshal(errResponse)
	if err != nil {
		// JSON marshalling of errors should always work.
		// If it fails then it might be some coding error we made ourselves.
		log.Fatalf("[error] [client %s]: Marshaling error response failed - %v", c.RemoteAddr(), err)
	}

	_, err = c.Write(j)
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
	}
}

func respondOK(c net.Conn) {
	msg := map[string]string{
		"status": "ok",
	}
	j, _ := json.Marshal(msg)
	_, err := c.Write(j)
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
	}
}
