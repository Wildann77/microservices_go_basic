package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/microservices-go/shared/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Client wraps RabbitMQ connection
type Client struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// Event represents a domain event
type Event struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
	Service   string          `json:"service"`
	TraceID   string          `json:"trace_id"`
}

// NewClient creates a new RabbitMQ client
func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return &Client{
		conn:    conn,
		channel: ch,
		url:     url,
	}, nil
}

// Close closes the connection
func (c *Client) Close() error {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// DeclareExchange declares a topic exchange
func (c *Client) DeclareExchange(name string) error {
	return c.channel.ExchangeDeclare(
		name,    // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
}

// DeclareQueue declares a queue
func (c *Client) DeclareQueue(name string) (amqp.Queue, error) {
	return c.channel.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

// BindQueue binds queue to exchange with routing key
func (c *Client) BindQueue(queue, exchange, routingKey string) error {
	return c.channel.QueueBind(
		queue,      // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,      // no-wait
		nil,        // args
	)
}

// Publish publishes an event
func (c *Client) Publish(ctx context.Context, exchange, routingKey string, event *Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	return c.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
			Headers: amqp.Table{
				"trace_id": event.TraceID,
				"service":  event.Service,
			},
		},
	)
}

// Consume starts consuming messages from a queue
func (c *Client) Consume(queue string) (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		queue, // queue
		"",    // consumer
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
}

// Publisher handles event publishing
type Publisher struct {
	client      *Client
	exchange    string
	serviceName string
}

// NewPublisher creates a new event publisher
func NewPublisher(client *Client, exchange, serviceName string) *Publisher {
	return &Publisher{
		client:      client,
		exchange:    exchange,
		serviceName: serviceName,
	}
}

// PublishEvent publishes a domain event
func (p *Publisher) PublishEvent(ctx context.Context, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	event := &Event{
		Type:      eventType,
		Payload:   data,
		Timestamp: time.Now(),
		Service:   p.serviceName,
		TraceID:   logger.GetTraceID(ctx),
	}

	routingKey := fmt.Sprintf("%s.%s", p.serviceName, eventType)

	logger.WithContext(ctx).Infof("Publishing event: %s", eventType)

	return p.client.Publish(ctx, p.exchange, routingKey, event)
}

// Consumer handles event consumption
type Consumer struct {
	client   *Client
	handlers map[string]func(context.Context, *Event) error
}

// NewConsumer creates a new event consumer
func NewConsumer(client *Client) *Consumer {
	return &Consumer{
		client:   client,
		handlers: make(map[string]func(context.Context, *Event) error),
	}
}

// RegisterHandler registers an event handler
func (c *Consumer) RegisterHandler(eventType string, handler func(context.Context, *Event) error) {
	c.handlers[eventType] = handler
}

// Start starts consuming events
func (c *Consumer) Start(queue string) error {
	msgs, err := c.client.Consume(queue)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	go func() {
		for msg := range msgs {
			var event Event
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				logger.Error("Failed to unmarshal event: " + err.Error())
				msg.Nack(false, false)
				continue
			}

			ctx := logger.SetTraceID(context.Background(), event.TraceID)

			handler, ok := c.handlers[event.Type]
			if !ok {
				logger.WithContext(ctx).Warnf("No handler for event type: %s", event.Type)
				msg.Ack(false)
				continue
			}

			if err := handler(ctx, &event); err != nil {
				logger.WithContext(ctx).WithError(err).Error("Failed to handle event")
				msg.Nack(false, true) // requeue
				continue
			}

			msg.Ack(false)
		}
	}()

	return nil
}

// Common event types
const (
	EventUserCreated    = "user.created"
	EventUserUpdated    = "user.updated"
	EventUserDeleted    = "user.deleted"
	EventOrderCreated   = "order.created"
	EventOrderUpdated   = "order.updated"
	EventOrderCancelled = "order.cancelled"
	EventPaymentSuccess = "payment.success"
	EventPaymentFailed  = "payment.failed"
)
