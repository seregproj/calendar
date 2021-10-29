package rbmq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/seregproj/calendar/internal/messagebroker"
	"github.com/streadway/amqp"
)

type NotificationsRBMQClient struct {
	ch           *amqp.Channel
	exchangeName string
	exchangeType string
	queueName    string
	routingKey   string
	deliveryMode uint8
}

func NewNotificationsRBMQClient() *NotificationsRBMQClient {
	return &NotificationsRBMQClient{
		exchangeName: "notifications",
		exchangeType: amqp.ExchangeDirect,
		queueName:    "queue1",
		routingKey:   "email",
		deliveryMode: amqp.Persistent,
	}
}

func (rc *NotificationsRBMQClient) Connect(ctx context.Context, dsn string) error {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return fmt.Errorf("cant create conn: %w", err)
	}

	rc.ch, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("cant open channel: %w", err)
	}

	go func() {
		defer rc.ch.Close()

		<-ctx.Done()
	}()

	if err = rc.ch.ExchangeDeclare(rc.exchangeName, rc.exchangeType, true, false, false, false, nil); err != nil {
		return fmt.Errorf("cant exchange declare: %w", err)
	}

	q, err := rc.ch.QueueDeclare(rc.queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("cant queue declare: %w", err)
	}

	if err = rc.ch.QueueBind(q.Name, rc.routingKey, rc.exchangeName, false, nil); err != nil {
		return fmt.Errorf("cant bind queue: %v with err: %w", q.Name, err)
	}

	return nil
}

func (rc *NotificationsRBMQClient) PushNotification(notification *messagebroker.Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("cant marshal notification: %v with err: %w", notification, err)
	}

	if err = rc.ch.Publish(rc.exchangeName, rc.routingKey, false, false, amqp.Publishing{
		Type:         "content/json",
		Body:         data,
		DeliveryMode: rc.deliveryMode,
	}); err != nil {
		return fmt.Errorf("cant publish: %v with error: %w", data, err)
	}

	return nil
}

func (rc *NotificationsRBMQClient) ConsumeNotifications(ctx context.Context) (
	<-chan messagebroker.Notification,
	error) {
	deliveries, err := rc.ch.Consume(rc.queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("start consuming: %w", err)
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
