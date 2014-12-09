package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ripple/downstream"
	"github.com/ripple/upstream"
)

func main() {
	// read from config
	hubID := "001"
	broker := "tcp://128.199.132.229:60000"
	network := "tcp4"
	laddr := ":8000"

	// connect to upstream
	up := upstream.New()
	go up.Connect(hubID, broker)

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
