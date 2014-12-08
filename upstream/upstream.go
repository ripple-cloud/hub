package upstream

import "github.com/ripple/message"

// Upstream defines the interface Ripple Hub uses to communicate with its upstream.
type Upstream interface {
	Connect() error

	Register(msg *message.Message) error
	Publish(msg *message.Message) error
	Deregister(msg *message.Message) error
}
