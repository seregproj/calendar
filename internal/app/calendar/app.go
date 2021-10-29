package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/seregproj/calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
	uuIDGen UUIDGenerator
}

type UUIDGenerator interface {
	Generate() (string, error)
}

type Logger interface {
	Warning(string)
	WarningWithFields(text string, fields map[string]interface{})
}

type Storage interface {
	ExistsEventByID(context.Context, string) (bool, error)
	CreateEvent(context.Context, *storage.Event) error
	UpdateEvent(context.Context, string, *storage.Event) error
	DeleteEvent(context.Context, string) error
	GetEventsByDaySorted(context.Context, time.Time, int64, int64) ([]*storage.Event, error)
}

var (
	ErrUnexpected         = errors.New("unknown error")
	ErrEventAlreadyExists = errors.New("event already exists")
	ErrEventNotFound      = errors.New("event not found")
	ErrInvalidDateFormat  = errors.New("invalid date format")
)

func New(logger Logger, storage Storage, uuidGen UUIDGenerator) *App {
	return &App{logger: logger, storage: storage, uuIDGen: uuidGen}
}

func (a *App) checkEventExists(ctx context.Context, uuid string) error {
	exists, err := a.storage.ExistsEventByID(ctx, uuid)
	if err != nil {
		a.logger.WarningWithFields(fmt.Sprintf("cant check exists event with err: %v", err.Error()), map[string]interface{}{
			"eventUUID": uuid,
		})

		return ErrUnexpected
	}

	if !exists {
		return ErrEventNotFound
	}

	return nil
}

func (a *App) CreateEvent(ctx context.Context, event *storage.Event) (string, error) {
	uuid, err := a.uuIDGen.Generate()
	if err != nil {
		a.logger.Warning(fmt.Sprintf("cant generate uuid: %v", err.Error()))

		return "", ErrUnexpected
	}

	event.ID = uuid

	err = a.storage.CreateEvent(ctx, event)
	if err != nil {
		a.logger.WarningWithFields(fmt.Sprintf("cant create event with err: %v", err.Error()), map[string]interface{}{
			"event": event,
		})

		return "", ErrUnexpected
	}

	return uuid, nil
}

func (a *App) UpdateEvent(ctx context.Context, uuid string, event *storage.Event) error {
	err := a.checkEventExists(ctx, uuid)
	if err != nil {
		return err
	}

	err = a.storage.UpdateEvent(ctx, uuid, event)
	if err != nil {
		a.logger.WarningWithFields(fmt.Sprintf("cant update event with err: %v", err.Error()), map[string]interface{}{
			"eventUUID": uuid,
		})

		return ErrUnexpected
	}

	return nil
}

func (a *App) DeleteEvent(ctx context.Context, uuid string) error {
	err := a.checkEventExists(ctx, uuid)
	if err != nil {
		return err
	}

	err = a.storage.DeleteEvent(ctx, uuid)
	if err != nil {
		a.logger.WarningWithFields(fmt.Sprintf("cant delete event with err: %v", err.Error()), map[string]interface{}{
			"eventUUID": uuid,
		})

		return ErrUnexpected
	}

	return nil
}

func (a *App) GetEventsByDay(ctx context.Context, day string, limit, offset int64) ([]*storage.Event, error) {
	dayTime, err := time.Parse("2006-01-02", day)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	events, err := a.storage.GetEventsByDaySorted(ctx, dayTime, limit, offset)
	if err != nil {
		a.logger.WarningWithFields(fmt.Sprintf("cant get events by day with err: %v", err.Error()), map[string]interface{}{
			"day": day,
		})

		return nil, ErrUnexpected
	}

	return events, nil
}
