package adapter

import (
	"errors"

	"github.com/joaohgf/mjv-challenge/internal/inbound/consumer/dto"
)

// validateMessage ensures an order handler only maps a complete message envelope.
func validateMessage(message *dto.Message[*dto.Order]) error {
	if message == nil {
		return errors.New("invalid message: envelope is required")
	}
	if message.Payload == nil {
		return errors.New("invalid message: payload is required")
	}
	return nil
}
