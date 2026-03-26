package consumer

// Consumer keeps outbound call dependencies for ADRF.
//
// This placeholder keeps free5gc-style layering and avoids refactoring package
// wiring once callback/notification flows are added.
type Consumer struct{}

func NewConsumer() (*Consumer, error) {
	return &Consumer{}, nil
}
