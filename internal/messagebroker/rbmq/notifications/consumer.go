package notifications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/seregproj/calendar/internal/messagebroker"
	"github.com/streadway/amqp"
)

type Consumer struct {
	ch    *amqp.Channel
	queue string
}

func NewConsumer(queue string) *Consumer {
	return &Consumer{
		queue: queue,
	}
}

func (c *Consumer) Connect(ctx context.Context, dsn string) error {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return fmt.Errorf("cant create conn: %w", err)
	}

	c.ch, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("cant open channel: %w", err)
	}

	go func() {
		defer c.ch.Close()

		<-ctx.Done()
	}()

	_, err = c.ch.QueueDeclare(c.queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("cant queue declare: %w", err)
	}

	return nil
}

func (c *Consumer) ConsumeNotifications(ctx context.Context) (
	<-chan messagebroker.Notification,
	error) {
	deliveries, err := c.ch.Consume(c.queue, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("cant consume with err %w", err)
	}

	notifications := make(chan messagebroker.Notification)
	go func() {
		defer close(notifications)

		for {
			select {
			case <-ctx.Done():
				return
			case d := <-deliveries:
				notification := messagebroker.Notification{}
				if err := json.Unmarshal(d.Body, &notification); err != nil {
					fmt.Printf("cant unmarshal obj from queue with err: %s\n", err.Error())

					continue
				}

				select {
				case <-ctx.Done():
					return
				case notifications <- notification:
					fmt.Println("sent to channel")
					if err := d.Ack(false); err != nil {
						fmt.Printf("cant ack with err: %s\n", err.Error())
					}
				}
			}
		}
	}()

	return notifications, nil
}
