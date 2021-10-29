package memorystorage_test

import (
	"context"
	"testing"
	"time"

	"github.com/seregproj/calendar/internal/app/calendar"
	"github.com/seregproj/calendar/internal/storage"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestCreateEvent(t *testing.T) {
	const uuid = "test"

	t.Run("test adding to empty/not-empty map", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		// adding to empty map
		expEvent := storage.Event{ID: uuid, Start: now, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)
		event, err := s.GetEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, &expEvent, event)

		// adding to not-empty map
		uuid2 := "test2"
		expEvent2 := storage.Event{ID: uuid2, Start: now, Finish: after, Description: "desc2", Title: "title2"}
		err = s.CreateEvent(ctx, &expEvent2)
		require.NoError(t, err)
		event2, err := s.GetEventByID(ctx, expEvent2.ID)
		require.NoError(t, err)
		require.Equal(t, &expEvent2, event2)
	})

	t.Run("test repeat adding", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		expEvent := storage.Event{ID: uuid, Start: now, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)
		event, err := s.GetEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, &expEvent, event)

		err = s.CreateEvent(ctx, &expEvent)
		require.ErrorIs(t, err, calendar.ErrEventAlreadyExists)
	})
}
