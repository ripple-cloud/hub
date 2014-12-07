package main

import (
	"errors"
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
	req, err := message.Decode(c)
	if err != nil {
		respondError(c, errors.New("invalid_request"))
		return
	}

	switch req.Type {
	case message.Register:
		err := up.Register(req)
		if err != nil {
			respondError(c, err)
			return
		}
		respondAck(c)
	case message.Publish:
		err := up.Publish(req)
		if err != nil {
			respondError(c, err)
			return
		}
		respondAck(c)
	case message.Deregister:
		err := up.Deregister(req)
		if err != nil {
			respondError(c, err)
			return
		}
		respondAck(c)
	default:
		respondError(c, errors.New("unknown_message_type"))
		return
	}
}

func respondError(c net.Conn, err error) {
	res := message.NewMessage()
	res.Type = message.Error
	res.Meta["error"] = err.Error()
	b, err := res.Encode()
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
		return
	}

	_, err = c.Write(b)
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
	}
}

func respondAck(c net.Conn) {
	res := message.NewMessage()
	res.Type = message.Ack
	b, err := res.Encode()
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
		return
	}

	_, err = c.Write(b)
	if err != nil {
		log.Printf("[error] [client %s]: %v", c.RemoteAddr(), err)
	}
}
