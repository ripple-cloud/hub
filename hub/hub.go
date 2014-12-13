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
	// TODO: read from config and handle flags
	hubID := "ripple-hub-001"
	broker := "tcp://128.199.132.229:60000"
	network := "tcp4"
	laddr := "0.0.0.0:8000"

	// connect to upstream
	up := upstream.NewMQTTUpstream()
	err := up.Connect(broker, hubID)
	if err != nil {
		panic(err)
	}
	defer up.Disconnect()

	log.Printf("[info] connected to Ripple Cloud as %s", hubID)

	// start listening to requests coming from downstream
	// using the given listener interface
	go func() {
		err := downstream.Listen(network, laddr, up)
		if err != nil {
			panic(err)
		}
	}()

	// Handle signals (CTRL-C)
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("[info] %v", <-ch)
}
