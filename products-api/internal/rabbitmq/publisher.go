package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"products-api/internal/config"
	"products-api/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Publisher sends product events to RabbitMQ.
type Publisher struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
}

// NewPublisher configures the RabbitMQ exchange and channel.
func NewPublisher(cfg *config.Config) (*Publisher, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("create channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		cfg.RabbitMQExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &Publisher{
		conn:     conn,
		channel:  ch,
		exchange: cfg.RabbitMQExchange,
	}, nil
}

// Close releases RabbitMQ resources.
func (p *Publisher) Close() {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.conn != nil {
		_ = p.conn.Close()
	}
}

// PublishProductCreated emits product.created.
func (p *Publisher) PublishProductCreated(product *models.Product) error {
	return p.publish("product.created", productPayload(product))
}

// PublishProductUpdated emits product.updated.
func (p *Publisher) PublishProductUpdated(product *models.Product) error {
	return p.publish("product.updated", productPayload(product))
}

// PublishProductDeleted emits product.deleted.
func (p *Publisher) PublishProductDeleted(id string) error {
	payload := map[string]string{"id": id}
	return p.publish("product.deleted", payload)
}

func (p *Publisher) publish(routingKey string, payload interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	return p.channel.PublishWithContext(
		ctx,
		p.exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now().UTC(),
		},
	)
}

func productPayload(p *models.Product) map[string]interface{} {
	if p == nil {
		return map[string]interface{}{}
	}
	id := ""
	if p.ID != primitive.NilObjectID {
		id = p.ID.Hex()
	}
	return map[string]interface{}{
		"id":          id,
		"name":        p.Name,
		"descripcion": p.Descripcion,
		"precio":      p.Precio,
		"stock":       p.Stock,
		"tipo":        p.Tipo,
		"estacion":    p.Estacion,
		"ocasion":     p.Ocasion,
		"notas":       p.Notas,
		"genero":      p.Genero,
		"marca":       p.Marca,
		"imagen":      p.Imagen,
		"owner_id":    p.OwnerID,
		"score":       p.Score,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,
	}
}
