package scheduler

import (
	"errors"
	"fmt"

	"github.com/seregproj/calendar/internal/messagebroker"
	"github.com/seregproj/calendar/internal/storage"
	"golang.org/x/net/context"
)

var ErrUnexpected = errors.New("unexpected error")

type App struct {
	logger  Logger
	storage Storage
	broker  MessageBroker
}

func New(logger Logger, storage Storage, broker MessageBroker) *App {
	return &App{logger: logger, storage: storage, broker: broker}
}

type Logger interface {
	Warning(text string)
	WarningWithFields(text string, fields map[string]interface{})
}

type Storage interface {
	GetUnprocessedActualEvents(ctx context.Context, limit int64) ([]*storage.Event, error)
	UpdateEventAsProcessed(ctx context.Context, event *storage.Event) error
}

type MessageBroker interface {
	PushNotification(notification *messagebroker.Notification) error
}

func (app *App) ProcessActualEvents(ctx context.Context, limit int64) error {
	events, err := app.storage.GetUnprocessedActualEvents(ctx, limit)
	if err != nil {
		app.logger.Warning(fmt.Sprintf("cant get unprocessed actual events with err: %v", err.Error()))

		return ErrUnexpected
	}

	for _, event := range events {
		if err := app.broker.PushNotification(messagebroker.NewNotification(event.ID, event.Title, event.Start)); err != nil {
			app.logger.WarningWithFields(fmt.Sprintf("cant push event to message broker: %v", err), map[string]interface{}{
				"event": event,
			})

			continue
		}

		if err := app.storage.UpdateEventAsProcessed(ctx, event); err != nil {
			app.logger.WarningWithFields(fmt.Sprintf("cant set event as processed: %v", err), map[string]interface{}{
				"event": event,
			})
		}
	}

	return nil
}
