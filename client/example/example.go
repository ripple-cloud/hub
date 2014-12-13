package main

import (
	"log"
	"net"

	"github.com/ripple/client"
	"github.com/ripple/message"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:60002")
	if err != nil {
		log.Fatalf("[error] %s", err)
	}
	defer l.Close()

	log.Printf("[info] Listening for %s on %s", l.Addr().Network(), l.Addr().String())

	// FIXME should get config from environment
	rc := client.New("tcp", "0.0.0.0:8000")

	// register with Ripple Hub
	req := message.NewRegister("example-echo", map[string]string{
		"network": l.Addr().Network(),
		"address": l.Addr().String(),
	})
	res, err := rc.Send(req)
	if err != nil {
		log.Fatal("[error] failed to dial Ripple Hub - %s", err)
	}
	if res.Type == message.Error {
		log.Fatalf("[error] failed to register with Ripple Cloud - %s", res.Meta["error"])
	}

	for {
		// accept requests and process them
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			req, err := message.Decode(conn)
			if err != nil {
				log.Printf("[error] %s", err)
			}

			if req.Type == message.Request {
				log.Printf("Received message: %s", req.Body)

				// publish the message
				req := message.NewPublish("example-echo", map[string]string{}, req.Body)
				res, err := rc.Send(req)
				if err != nil {
					log.Printf("[error] %s", err)
				}
				if res.Type == message.Error {
					log.Printf("[error] %s", err)
				}
			}
		}()
	}
}
