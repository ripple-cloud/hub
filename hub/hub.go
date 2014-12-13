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
	hubID := "ripple-hub-001"
	broker := "tcp://128.199.132.229:60000"
	network := "tcp4"
	laddr := ":8000"

	// connect to upstream
	up := upstream.NewMQTTUpstream()
	go func() {
		defer up.Disconnect()
		err := up.Connect(broker, hubID)
		if err != nil {
			panic(err)
		}
	}()

	// start listening to requests coming from downstream
	// using the given listener interface
	go func() {
		defer downstream.Close()
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
