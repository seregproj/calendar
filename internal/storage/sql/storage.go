package sqlstorage

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/seregproj/calendar/internal/storage"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Connect(ctx context.Context, dsn string) error {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return err
	}

	s.pool = pool

	return nil
}

func (s *Storage) Close(ctx context.Context) {
	s.pool.Close()
}

func (s *Storage) ExistsEventByID(ctx context.Context, uuid string) (bool, error) {
	ct, err := s.pool.Exec(ctx, "SELECT 1 FROM events where id = $1", uuid)
	if err != nil {
		return false, fmt.Errorf("cant exec: %w", err)
	}

	return ct.RowsAffected() > 0, nil
}

func (s *Storage) CreateEvent(ctx context.Context, event *storage.Event) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO events(id, title, description, datetime_start, datetime_finish) "+
		"VALUES ($1, $2, $3, $4, $5)", event.ID, event.Title, event.Description, event.Start, event.Finish)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	return nil
}

func (s *Storage) UpdateEvent(ctx context.Context, uuid string, event *storage.Event) error {
	_, err := s.pool.Exec(ctx, "UPDATE events SET title=$1, description=$2, datetime_start=$3, datetime_finish=$4"+
		"  WHERE id=$5", event.Title, event.Description, event.Start, event.Finish, uuid)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	return nil
}

func (s *Storage) DeleteEvent(ctx context.Context, uuid string) error {
	_, err := s.pool.Exec(ctx, "DELETE FROM events WHERE id=$1", uuid)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	return nil
}

func (s *Storage) GetEventsByDaySorted(ctx context.Context, date time.Time, limit int64, offset int64) (
	[]*storage.Event,
	error) {
	dateTo := date.AddDate(0, 0, 1)

	var eventsDB []Event
	if err := pgxscan.Select(ctx, s.pool, &eventsDB,
		"SELECT * FROM events where datetime_start >= $1 AND datetime_start < $2 LIMIT $3 OFFSET $4",
		date, dateTo, limit, offset); err != nil {
		return nil, fmt.Errorf("cant do select: %w", err)
	}

	events := make([]*storage.Event, 0, len(eventsDB))
	for _, item := range eventsDB {
		event := item.ToApp()
		events = append(events, &event)
	}

	return events, nil
}

func (s *Storage) GetUnprocessedActualEvents(ctx context.Context, limit int64) ([]*storage.Event, error) {
	var eventsDB []Event
	if err := pgxscan.Select(ctx, s.pool, &eventsDB,
		"SELECT * FROM events where datetime_start < $1 AND NOT processed LIMIT $2", time.Now(), limit); err != nil {
		return nil, fmt.Errorf("cant do select: %w", err)
	}

	events := make([]*storage.Event, 0, len(eventsDB))

	for _, item := range eventsDB {
		event := item.ToApp()
		events = append(events, &event)
	}

	return events, nil
}

func (s *Storage) UpdateEventAsProcessed(ctx context.Context, event *storage.Event) error {
	_, err := s.pool.Exec(ctx, "UPDATE events SET processed=true WHERE id=$1", event.ID)
	if err != nil {
		return fmt.Errorf("exec error: %w", err)
	}

	return nil
}
