package sender

import (
	"context"
	"errors"
	"fmt"

	"github.com/seregproj/calendar/internal/messagebroker"
)

var ErrUnexpected = errors.New("unexpected error")

type App struct {
	logger Logger
	broker MessageBroker
}

func New(logger Logger, broker MessageBroker) *App {
	return &App{logger: logger, broker: broker}
}

type Logger interface {
	Info(text string)
	Warning(text string)
	WarningWithFields(text string, fields map[string]interface{})
}

type MessageBroker interface {
	ConsumeNotifications(ctx context.Context) (<-chan messagebroker.Notification, error)
}

func (app *App) SendNotifications(ctx context.Context) error {
	notifs, err := app.broker.ConsumeNotifications(ctx)
	if err != nil {
		app.logger.Warning(fmt.Sprintf("cant consume notifs with err: %v", err))

		return ErrUnexpected
	}

	app.logger.Info("ready to send notifications..")
	for n := range notifs {
		app.logger.Info(fmt.Sprintf("sent notif: %v", n))
	}

	return nil
}
