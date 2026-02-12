package rabbit

import (
	"context"
	"encoding/json"

	"github.com/microservices-go/services/order/internal/order"
	"github.com/microservices-go/shared/logger"
	"github.com/microservices-go/shared/rabbitmq"
)

// Consumer handles RabbitMQ events
type Consumer struct {
	service *order.Service
}

// NewConsumer creates a new consumer
func NewConsumer(service *order.Service) *Consumer {
	return &Consumer{service: service}
}

// Start starts consuming events
func (c *Consumer) Start(client *rabbitmq.Client) error {
	// Declare queue
	queue, err := client.DeclareQueue("order-service-queue")
	if err != nil {
		return err
	}

	// Bind to exchange with routing keys
	if err := client.BindQueue(queue.Name, "microservices.events", "user-service.user.created"); err != nil {
		return err
	}
	if err := client.BindQueue(queue.Name, "microservices.events", "payment-service.payment.success"); err != nil {
		return err
	}
	if err := client.BindQueue(queue.Name, "microservices.events", "payment-service.payment.failed"); err != nil {
		return err
	}

	// Create consumer
	consumer := rabbitmq.NewConsumer(client)

	// Register handlers
	consumer.RegisterHandler("user.created", c.handleUserCreated)
	consumer.RegisterHandler("payment.success", c.handlePaymentSuccess)
	consumer.RegisterHandler("payment.failed", c.handlePaymentFailed)

	return consumer.Start(queue.Name)
}

func (c *Consumer) handleUserCreated(ctx context.Context, event *rabbitmq.Event) error {
	log := logger.WithContext(ctx)
	log.Info("Handling user.created event")

	var payload struct {
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return c.service.HandleUserCreated(ctx, payload.UserID)
}

func (c *Consumer) handlePaymentSuccess(ctx context.Context, event *rabbitmq.Event) error {
	log := logger.WithContext(ctx)
	log.Info("Handling payment.success event")

	var payload struct {
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return c.service.HandlePaymentSuccess(ctx, payload.OrderID)
}

func (c *Consumer) handlePaymentFailed(ctx context.Context, event *rabbitmq.Event) error {
	log := logger.WithContext(ctx)
	log.Info("Handling payment.failed event")

	var payload struct {
		OrderID string `json:"order_id"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return err
	}

	return c.service.HandlePaymentFailed(ctx, payload.OrderID)
}
