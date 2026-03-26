package consumer

// Consumer keeps outbound call dependencies for ADRF.
//
// R01 includes this placeholder to match free5gc layering and to avoid
// refactoring package wiring once ADRF starts callback/notification flows.
type Consumer struct{}

func NewConsumer() (*Consumer, error) {
	return &Consumer{}, nil
}
