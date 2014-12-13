package client

import (
	"bufio"
	"bytes"
	"net"

	"github.com/ripple-cloud/hub/message"
)

type RippleClient struct {
	Network string
	Addr    string
}

func New(network, addr string) *RippleClient {
	return &RippleClient{
		Network: network,
		Addr:    addr,
	}
}

func (c *RippleClient) Send(req *message.Message) (*message.Message, error) {
	// dial tcp address
	conn, err := net.Dial(c.Network, c.Addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// send message
	b, err := req.Encode()
	if err != nil {
		return nil, err
	}
	if _, err = conn.Write(b); err != nil {
		return nil, err
	}

	// read the response
	b, err = bufio.NewReader(conn).ReadBytes('\n')
	if b == nil && err != nil {
		return nil, err
	}

	return message.Decode(bytes.NewReader(b))
}
