package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	// EventProductCreated is emitted when a product is created.
	EventProductCreated = "product.created"
	// EventProductUpdated is emitted when a product is updated.
	EventProductUpdated = "product.updated"
	// EventProductDeleted is emitted when a product is deleted.
	EventProductDeleted = "product.deleted"
)

// ProductEvent captures the routing information flowing through RabbitMQ.
type ProductEvent struct {
	Type      string
	ProductID string
}

// EventHandler processes product events.
type EventHandler interface {
	HandleProductEvent(ctx context.Context, event ProductEvent) error
}

// ConsumerConfig holds connection parameters.
type ConsumerConfig struct {
	URL      string
	Exchange string
	Queue    string
}

// Consumer consumes product events and forwards them to an EventHandler.
type Consumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queue    string
	exchange string
	handler  EventHandler
}

// NewConsumer connects to RabbitMQ, declares exchange/queue/bindings and returns a Consumer.
func NewConsumer(cfg ConsumerConfig, handler EventHandler) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("create channel: %w", err)
	}

	if err := ch.ExchangeDeclare(cfg.Exchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	queue, err := ch.QueueDeclare(cfg.Queue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	bindings := []string{EventProductCreated, EventProductUpdated, EventProductDeleted}
	for _, key := range bindings {
		if err := ch.QueueBind(queue.Name, key, cfg.Exchange, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return nil, fmt.Errorf("bind %s: %w", key, err)
		}
	}

	return &Consumer{
		conn:     conn,
		channel:  ch,
		queue:    queue.Name,
		exchange: cfg.Exchange,
		handler:  handler,
	}, nil
}

// Start begins consuming messages until the context is canceled or the channel closes.
func (c *Consumer) Start(ctx context.Context) error {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("channel closed")
			}

			event, err := decodeProductEvent(msg.RoutingKey, msg.Body)
			if err != nil {
				log.Printf("skip message (decode error): %v", err)
				_ = msg.Ack(false)
				continue
			}

			if err := c.handler.HandleProductEvent(ctx, event); err != nil {
				log.Printf("handler error: %v", err)
			}

			_ = msg.Ack(false)
		}
	}
}

// Close releases underlying AMQP resources.
func (c *Consumer) Close() {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

func decodeProductEvent(routingKey string, body []byte) (ProductEvent, error) {
	eventType, err := normalizeRoutingKey(routingKey)
	if err != nil {
		return ProductEvent{}, err
	}

	var payload productMessage
	if err := json.Unmarshal(body, &payload); err != nil {
		return ProductEvent{}, fmt.Errorf("decode payload: %w", err)
	}

	return ProductEvent{
		Type:      eventType,
		ProductID: payload.ID,
	}, nil
}

func normalizeRoutingKey(routingKey string) (string, error) {
	switch routingKey {
	case EventProductCreated, EventProductUpdated, EventProductDeleted:
		return routingKey, nil
	default:
		return "", fmt.Errorf("unknown routing key %s", routingKey)
	}
}

type productMessage struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Descripcion string     `json:"descripcion"`
	Precio      float64    `json:"precio"`
	Stock       int        `json:"stock"`
	Tipo        string     `json:"tipo"`
	Estacion    string     `json:"estacion"`
	Ocasion     string     `json:"ocasion"`
	Notas       []string   `json:"notas"`
	Genero      string     `json:"genero"`
	Marca       string     `json:"marca"`
	Imagen      string     `json:"imagen"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
