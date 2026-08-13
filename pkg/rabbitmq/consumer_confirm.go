package adapter

// confirmationEnabler enables RabbitMQ publisher-confirm mode on a channel.
type confirmationEnabler interface {
	Confirm(bool) error
}

// enableConfirms requires RabbitMQ to confirm each publication accepted by the broker.
func enableConfirms(channel confirmationEnabler) error {
	return channel.Confirm(false)
}
