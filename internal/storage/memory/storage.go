package memorystorage

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/seregproj/calendar/internal/app/calendar"
	"github.com/seregproj/calendar/internal/storage"
)

type Storage struct {
	sync.RWMutex
	events map[string]*Event
}

func New() *Storage {
	return &Storage{
		events: make(map[string]*Event),
	}
}

func (s *Storage) ExistsEventByID(ctx context.Context, uuid string) (bool, error) {
	s.RLock()
	defer s.RUnlock()

	_, ok := s.events[uuid]

	return ok, nil
}

func (s *Storage) GetEventByID(ctx context.Context, uuid string) (*storage.Event, error) {
	s.RLock()
	defer s.RUnlock()

	e, ok := s.events[uuid]

	if !ok {
		return nil, calendar.ErrEventNotFound
	}

	appEvent := e.ToApp()
	return &appEvent, nil
}

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.events[event.ID]; ok {
		return calendar.ErrEventAlreadyExists
	}

	s.events[event.ID] = NewFromApp(event)

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, uuid string, event *storage.Event) error {
	s.Lock()
	defer s.Unlock()

	e, ok := s.events[uuid]

	if !ok {
		return calendar.ErrEventNotFound
	}

	e.UpdateFromApp(event)
	s.events[uuid] = e

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, uuid string) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.events[uuid]

	if !ok {
		return calendar.ErrEventNotFound
	}

	delete(s.events, uuid)

	return nil
}

func (s *Storage) GetEventsByDaySorted(ctx context.Context, date time.Time, limit, offset int64) (
	[]*storage.Event,
	error) {
	s.RLock()
	defer s.RUnlock()

	events := make([]*storage.Event, 0, limit)
	dateFrom, err := time.Parse("2006-01-02", date.Format("2006-01-02"))
	if err != nil {
		return events, fmt.Errorf("cant parse date: %w", err)
	}

	dateTo := dateFrom.AddDate(0, 0, 1)

	sortedEvents := make([]*Event, 0, len(s.events))
	for _, v := range s.events {
		select {
		case <-ctx.Done():
			return events, nil
		default:
		}

		sortedEvents = append(sortedEvents, v)
	}

	sort.Slice(sortedEvents, func(i, j int) bool {
		return sortedEvents[i].DatetimeStart.Before(sortedEvents[j].DatetimeStart)
	})

	for _, v := range sortedEvents {
		select {
		case <-ctx.Done():
			return events, nil
		default:
		}

		if dateFrom.Before(v.DatetimeStart) && dateTo.After(v.DatetimeStart) {
			if offset > 0 {
				offset--

				continue
			}

			eventApp := v.ToApp()
			events = append(events, &eventApp)

			limit--
		}

		if limit == 0 {
			break
		}
	}

	return events, nil
}

func (s *Storage) GetUnprocessedActualEvents(ctx context.Context, limit int64) ([]*storage.Event, error) {
	s.RLock()
	defer s.RUnlock()

	events := make([]*storage.Event, 0, limit)

	for _, v := range s.events {
		select {
		case <-ctx.Done():
			return events, nil
		default:
		}

		if time.Now().Before(v.DatetimeStart) {
			continue
		}

		if !v.Processed {
			eventApp := v.ToApp()
			events = append(events, &eventApp)

			limit--
		}

		if limit == 0 {
			break
		}
	}

	return events, nil
}

func (s *Storage) UpdateEventAsProcessed(ctx context.Context, event *storage.Event) error {
	s.Lock()
	defer s.Unlock()

	e, ok := s.events[event.ID]

	if !ok {
		return calendar.ErrEventNotFound
	}

	e.Processed = true
	s.events[event.ID] = e

	return nil
}
