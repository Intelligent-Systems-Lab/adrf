package consumer

// Consumer keeps outbound call dependencies for ADRF.
type Consumer struct {
	nrfService *NrfService
}

func NewConsumer() (*Consumer, error) {
	return &Consumer{
		nrfService: NewNrfService(),
	}, nil
}

func (c *Consumer) NrfService() *NrfService {
	if c == nil {
		return nil
	}
	return c.nrfService
}
