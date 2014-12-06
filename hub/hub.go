package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ripple/message"
)

// Upstream defines the interface Ripple Hub uses to communicate with its upstream.
type Upstream interface {
	Connect() error

	Register(msg *message.Message) error
	Publish(msg *message.Message) error
	Deregister(msg *message.Message) error
}

const appName = "Ripple Hub"

func main() {
	// read from the config
	network := "tcp4"
	laddr := ":8000"
	up := newMQTTUpstream()

	// connect to upstream
	go up.Connect()

	// start listening to requests coming from downstream
	// using the given listener interface
	go listenDownstream(network, laddr, up)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("[info] %v", <-ch)

	// TODO
	// Disconnect from upstream and close downstream connections gracefully
}
