package downstream

import (
	"errors"
	"log"
	"net"

	"github.com/ripple-cloud/hub/message"
	"github.com/ripple-cloud/hub/upstream"
)

func Listen(network, laddr string, up upstream.Upstream) error {
	l, err := net.Listen(network, laddr)
	if err != nil {
		log.Fatalf("[error] %s", err)
	}
	defer l.Close()

	log.Printf("[info] Listening for %s on %s", network, l.Addr())

	for {
		// accept requests and process them
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("[error] %s", err)
		}

		go downstreamHandler(conn, up)
	}
}

// downstreamHandler will relay requests to upstream
func downstreamHandler(c net.Conn, up upstream.Upstream) {
	defer c.Close()

	// decode the incoming message
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
	res := message.New()
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
	res := message.New()
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
