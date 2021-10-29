package memorystorage_test

import (
	"context"
	"testing"
	"time"

	"github.com/seregproj/calendar/internal/storage"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestGetEventsByDaySortedEmpty(t *testing.T) {
	t.Run("test no events with empty map", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		now := time.Now()

		botd, err := time.Parse("2006-01-02", now.Format("2006-01-02"))
		require.NoError(t, err)
		events, err := s.GetEventsByDaySorted(ctx, botd, 1, 0)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})

	t.Run("test no events with not empty map", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		begin := time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC)
		after := begin.AddDate(0, 0, 2)

		event := storage.Event{ID: "test", Start: begin, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &event)
		require.NoError(t, err)

		botnd, err := time.Parse("2006-01-02", begin.AddDate(0, 0, 1).
			Format("2006-01-02"))
		require.NoError(t, err)
		events, err := s.GetEventsByDaySorted(ctx, botnd, 1, 0)
		require.NoError(t, err)
		require.Equal(t, 0, len(events))
	})
}

func TestGetEventsByDaySortedValid(t *testing.T) {
	validTests := map[string]struct {
		createEvents []*storage.Event
		dateSearch   time.Time
		limit        int64
		offset       int64
		expEvents    []*storage.Event
	}{
		"get events with limit 1 and offset 0": {
			createEvents: []*storage.Event{
				{
					ID:          "event1",
					Start:       time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc1",
					Title:       "title1",
				},
				{
					ID:          "event2",
					Start:       time.Date(2020, 10, 11, 23, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc2",
					Title:       "title2",
				},
				{
					ID:          "event3",
					Start:       time.Date(2020, 10, 12, 23, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc3",
					Title:       "title3",
				},
			},
			dateSearch: time.Date(2020, 10, 11, 0o0, 0o0, 0, 0, time.UTC),
			limit:      1,
			offset:     0,
			expEvents: []*storage.Event{
				{
					ID:          "event1",
					Start:       time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc1",
					Title:       "title1",
				},
			},
		},
		"get events with limit 1 and offset 2": {
			createEvents: []*storage.Event{
				{
					ID:          "event1",
					Start:       time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc1",
					Title:       "title1",
				},
				{
					ID:          "event2",
					Start:       time.Date(2020, 10, 11, 22, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc2",
					Title:       "title2",
				},
				{
					ID:          "event3",
					Start:       time.Date(2020, 10, 11, 23, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc3",
					Title:       "title3",
				},
				{
					ID:          "event4",
					Start:       time.Date(2020, 10, 11, 16, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc4",
					Title:       "title4",
				},
			},
			dateSearch: time.Date(2020, 10, 11, 0o0, 0o0, 0, 0, time.UTC),
			limit:      1,
			offset:     2,
			expEvents: []*storage.Event{
				{
					ID:          "event2",
					Start:       time.Date(2020, 10, 11, 22, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc2",
					Title:       "title2",
				},
			},
		},
		"no events with limit 1 and offset 1": {
			createEvents: []*storage.Event{
				{
					ID:          "event1",
					Start:       time.Date(2020, 10, 11, 15, 16, 0, 0, time.UTC),
					Finish:      time.Date(2020, 10, 13, 15, 16, 0, 0, time.UTC),
					Description: "desc1",
					Title:       "title1",
				},
			},
			dateSearch: time.Date(2020, 10, 11, 0o0, 0o0, 0, 0, time.UTC),
			limit:      1,
			offset:     1,
			expEvents:  []*storage.Event{},
		},
	}

	for testName, data := range validTests {
		s := memorystorage.New()
		ctx := context.Background()

		t.Run(testName, func(t *testing.T) {
			for _, item := range data.createEvents {
				err := s.CreateEvent(ctx, item)
				require.NoError(t, err)
			}

			events, err := s.GetEventsByDaySorted(ctx, data.dateSearch, data.limit, data.offset)
			require.NoError(t, err)
			require.Equal(t, data.expEvents, events)
		})
	}
}
