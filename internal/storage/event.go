package storage

import (
	"errors"
	"fmt"
	"time"
)

type Event struct {
	ID          string
	Title       string
	Description string
	Start       time.Time
	Finish      time.Time
}

var (
	ErrDatestartBeforeNow   = errors.New("datestart should be in future")
	ErrDatestartAfterFinish = errors.New("datestart should be before datefinish")
)

func NewEvent(id, title, description string, dateStart, dateFinish time.Time) (*Event, error) {
	if dateStart.Before(time.Now()) {
		return nil, fmt.Errorf("invalid datestart: %v, %w", dateStart, ErrDatestartBeforeNow)
	}

	if dateStart.After(dateFinish) {
		return nil, fmt.Errorf("invalid datestart: %v, %w", dateStart, ErrDatestartAfterFinish)
	}

	ds, err := time.Parse("2006-01-02 15:04", dateStart.Format("2006-01-02 15:04"))
	if err != nil {
		return nil, fmt.Errorf("cant parse datestart: %w", err)
	}

	df, err := time.Parse("2006-01-02 15:04", dateFinish.Format("2006-01-02 15:04"))
	if err != nil {
		return nil, fmt.Errorf("cant parse datefinish: %w", err)
	}

	return &Event{
		ID:          id,
		Title:       title,
		Description: description,
		Start:       ds,
		Finish:      df,
	}, nil
}
