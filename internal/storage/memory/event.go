package memorystorage

import (
	"time"

	"github.com/seregproj/calendar/internal/storage"
)

type Event struct {
	ID             string
	Title          string
	Description    string
	DatetimeStart  time.Time
	DatetimeFinish time.Time
	Processed      bool
}

func NewFromApp(e *storage.Event) *Event {
	event := Event{}
	event.ID = e.ID
	event.Title = e.Title
	event.Description = e.Description
	event.DatetimeStart = e.Start
	event.DatetimeFinish = e.Finish

	return &event
}

func (e *Event) ToApp() storage.Event {
	event := storage.Event{}
	event.ID = e.ID
	event.Title = e.Title
	event.Description = e.Description
	event.Start = e.DatetimeStart
	event.Finish = e.DatetimeFinish

	return event
}

func (e *Event) UpdateFromApp(event *storage.Event) {
	e.ID = event.ID
	e.Title = event.Title
	e.Description = event.Description
	e.DatetimeStart = event.Start
	e.DatetimeFinish = event.Finish
}
