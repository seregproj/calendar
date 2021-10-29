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

func TestDeleteEvent(t *testing.T) {
	const uuid = "test"

	t.Run("test delete existing event", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		// adding to empty map
		expEvent := storage.Event{ID: uuid, Start: now, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)

		err = s.DeleteEvent(ctx, expEvent.ID)
		require.NoError(t, err)
		_, err = s.GetEventByID(ctx, expEvent.ID)
		require.ErrorIs(t, err, calendar.ErrEventNotFound)
	})

	t.Run("test delete not-existing event", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		expEvent := storage.Event{ID: uuid, Start: now, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)

		notExisting := "not found"
		err = s.DeleteEvent(ctx, notExisting)
		require.ErrorIs(t, err, calendar.ErrEventNotFound)
		event, err := s.GetEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, &expEvent, event)
	})
}
