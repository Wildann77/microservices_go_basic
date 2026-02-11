package rabbit

import (
	"context"
	"encoding/json"

	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/rabbitmq"
	"github.com/microservices-go/services/payment/internal/payment"
)

// Consumer handles RabbitMQ events
type Consumer struct {
	service *payment.Service
}

// NewConsumer creates a new consumer
func NewConsumer(service *payment.Service) *Consumer {
	return &Consumer{service: service}
}

// Start starts consuming events
func (c *Consumer) Start(client *rabbitmq.Client) error {
	// Declare queue
	queue, err := client.DeclareQueue("payment-service-queue")
	if err != nil {
		return err
	}

	// Bind to exchange with routing keys
	if err := client.BindQueue(queue.Name, "microservices.events", "order-service.order.created"); err != nil {
		return err
	}

	// Create consumer
	consumer := rabbitmq.NewConsumer(client)

	// Register handlers
	consumer.RegisterHandler("order.created", c.handleOrderCreated)

	return consumer.Start(queue.Name)
}

func (c *Consumer) handleOrderCreated(ctx context.Context, event *rabbitmq.Event) error {
	log := logger.WithContext(ctx)
	log.Info("Handling order.created event")

	var payload struct {
		OrderID     string  `json:"order_id"`
		UserID      string  `json:"user_id"`
		TotalAmount float64 `json:"total_amount"`
		Currency    string  `json:"currency"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return c.service.HandleOrderCreated(ctx, payload.OrderID, payload.UserID, payload.TotalAmount, payload.Currency)
}