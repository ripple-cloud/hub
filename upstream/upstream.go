package upstream

import (
	"fmt"

	"github.com/ripple-cloud/common/message"
)

// Upstream defines the interface Ripple Hub uses to communicate with its upstream.
type Upstream interface {
	Connect(address, id string) error
	Disconnect()

	Register(msg *message.Message) error
	Publish(msg *message.Message) error
	Deregister(msg *message.Message) error
}

type RequiredFieldMissingError struct {
	Field string
}

func (e RequiredFieldMissingError) Error() string {
	return fmt.Sprintf("%s is required", e.Field)
}
