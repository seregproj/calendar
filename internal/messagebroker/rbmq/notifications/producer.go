package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/seregproj/calendar/internal/messagebroker"
	"github.com/streadway/amqp"
)

type Producer struct {
	ch         *amqp.Channel
	routingKey string
}

func NewProducer(routingKey string) *Producer {
	return &Producer{
		routingKey: routingKey,
	}
}

func (p *Producer) Connect(ctx context.Context, dsn string) error {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return fmt.Errorf("cant create conn: %w", err)
	}

	p.ch, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("cant open channel: %w", err)
	}

	go func() {
		defer p.ch.Close()

		<-ctx.Done()
	}()

	return nil
}

func (p *Producer) PushNotification(notification *messagebroker.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("cant marshal notification: %v with err: %w", notification, err)
	}

	if err = p.ch.Publish("", p.routingKey, false, false, amqp.Publishing{
		Type:         "content/json",
		Body:         data,
		DeliveryMode: amqp.Persistent,
	}); err != nil {
		return fmt.Errorf("cant publish: %v with error: %w", data, err)
	}

	return nil
}
