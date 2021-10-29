package memorystorage_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/seregproj/calendar/internal/storage"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestGetUnprocessedActualEvents(t *testing.T) {
	t.Run("test no events with empty map", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		events, err := s.GetUnprocessedActualEvents(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})

	t.Run("test no events with not empty map (have processed events only)", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		begin := time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC)
		after := begin.AddDate(0, 0, 2)

		event := storage.Event{ID: "test", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &event)
		require.NoError(t, err)
		err = s.UpdateEventAsProcessed(ctx, &event)
		require.NoError(t, err)

		events, err := s.GetUnprocessedActualEvents(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})

	t.Run("test no events with not empty map (have unprocessed but without actual start date event)", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		begin := time.Now().AddDate(0, 0, 1)
		after := begin.Add(time.Hour * 1)

		event := storage.Event{ID: "test", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &event)
		require.NoError(t, err)

		events, err := s.GetUnprocessedActualEvents(ctx, 1)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})

	t.Run("test get events with limit 2 (check events)", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()
		begin := time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC)
		after := begin.Add(time.Hour * 1)

		event1 := storage.Event{ID: "test1", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &event1)
		require.NoError(t, err)

		event2 := storage.Event{ID: "test2", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err = s.CreateEvent(ctx, &event2)
		require.NoError(t, err)
		err = s.UpdateEventAsProcessed(ctx, &event2)
		require.NoError(t, err)

		event3 := storage.Event{ID: "test3", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err = s.CreateEvent(ctx, &event3)
		require.NoError(t, err)

		events, err := s.GetUnprocessedActualEvents(ctx, 2)
		require.NoError(t, err)

		expSlice := []*storage.Event{
			&event1,
			&event3,
		}
		sort.Slice(events, func(i, j int) bool {
			return events[i].ID < events[j].ID
		})

		require.Equal(t, expSlice, events)
	})

	t.Run("test get events with limit 2 (check length)", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()
		begin := time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC)
		after := begin.Add(time.Hour * 1)

		event1 := storage.Event{ID: "test1", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &event1)
		require.NoError(t, err)

		event2 := storage.Event{ID: "test2", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err = s.CreateEvent(ctx, &event2)
		require.NoError(t, err)
		err = s.UpdateEventAsProcessed(ctx, &event2)
		require.NoError(t, err)

		event3 := storage.Event{ID: "test3", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err = s.CreateEvent(ctx, &event3)
		require.NoError(t, err)

		event4 := storage.Event{ID: "test4", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err = s.CreateEvent(ctx, &event4)
		require.NoError(t, err)

		events, err := s.GetUnprocessedActualEvents(ctx, 2)
		require.NoError(t, err)
		require.Equal(t, 2, len(events))
	})
}
