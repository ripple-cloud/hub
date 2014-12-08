package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ripple/downstream"
	"github.com/ripple/upstream"
)

const appName = "Ripple Hub"

func main() {
	// read from the config
	network := "tcp4"
	laddr := ":8000"
	up := upstream.NewMQTTUpstream()

	// connect to upstream
	go up.Connect()

	// start listening to requests coming from downstream
	// using the given listener interface
	go downstream.Listen(network, laddr, up)

	// Handle SIGINT and SIGTERM.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("[info] %v", <-ch)

	// TODO
	// Disconnect from upstream and close downstream connections gracefully
}
