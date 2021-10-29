package sqlstorage

import (
	"time"

	"github.com/seregproj/calendar/internal/storage"
)

type Event struct {
	ID             string    `db:"id"`
	Title          string    `db:"title"`
	Description    string    `db:"description"`
	DatetimeStart  time.Time `db:"datetime_start"`
	DatetimeFinish time.Time `db:"datetime_finish"`
	Processed      bool      `db:"processed"`
	DateAdd        time.Time `db:"date_add"`
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
