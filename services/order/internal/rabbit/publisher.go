package rabbit

import (
	"context"

	"github.com/microservices-go/shared/rabbitmq"
)

// Publisher wraps rabbitmq publisher for order service
type Publisher struct {
	publisher *rabbitmq.Publisher
}

// NewPublisher creates a new publisher
func NewPublisher(client *rabbitmq.Client) *Publisher {
	return &Publisher{
		publisher: rabbitmq.NewPublisher(client, "microservices.events", "order-service"),
	}
}

// PublishEvent publishes an event
func (p *Publisher) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	return p.publisher.PublishEvent(ctx, eventType, payload)
}
