package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewConsumer(url string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connecting to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("opening channel: %w", err)
	}

	return &Consumer{conn: conn, channel: ch}, nil
}

// Subscribe declara o exchange, a queue (com DLX configurado) e o binding, e inicia uma goroutine
// para consumir mensagens. Mensagens nackadas sem requeue são roteadas para <exchange>.dlx e
// armazenadas em <queue>.dlq. A goroutine encerra quando o channel for fechado (via Close).
func (c *Consumer) Subscribe(exchange, queue, routingKey string, handler func([]byte) error) error {
	dlx := exchange + ".dlx"
	dlq := queue + ".dlq"

	if err := c.channel.ExchangeDeclare(dlx, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declaring dead letter exchange %q: %w", dlx, err)
	}

	if _, err := c.channel.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declaring dead letter queue %q: %w", dlq, err)
	}

	if err := c.channel.QueueBind(dlq, "#", dlx, false, nil); err != nil {
		return fmt.Errorf("binding dlq %q to dlx %q: %w", dlq, dlx, err)
	}

	if err := c.channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declaring exchange %q: %w", exchange, err)
	}

	if _, err := c.channel.QueueDeclare(queue, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": dlx,
	}); err != nil {
		return fmt.Errorf("declaring queue %q: %w", queue, err)
	}

	if err := c.channel.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		return fmt.Errorf("binding queue %q to %q: %w", queue, routingKey, err)
	}

	msgs, err := c.channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("registering consumer on %q: %w", queue, err)
	}

	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				log.Printf("error handling %s: %v — nacking", msg.RoutingKey, err)
				msg.Nack(false, false)
				continue
			}
			msg.Ack(false)
		}
	}()

	return nil
}

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}
