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

func TestUpdateEvent(t *testing.T) {
	t.Run("test update existing event", func(t *testing.T) {
		ctx := context.Background()
		uuid := "test"
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		// adding to empty map
		expEvent := storage.Event{ID: uuid, Start: now, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)
		expEvent.Start = now.Add(time.Minute * 2)
		expEvent.Finish = now.Add(time.Minute * 4)
		expEvent.Description = "desc_updated"
		expEvent.Title = "title_updated"

		err = s.UpdateEvent(ctx, expEvent.ID, &expEvent)
		require.NoError(t, err)
		event, err := s.GetEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, &expEvent, event)
	})

	t.Run("test update not-existing event", func(t *testing.T) {
		ctx := context.Background()
		uuid := "test"
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		expEvent := storage.Event{ID: uuid, Start: now, Finish: after, Description: "desc1", Title: "title1"}
		err := s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)
		expEventU := storage.Event{
			ID: uuid, Start: now.Add(time.Minute * 2), Finish: now.Add(time.Minute * 4),
			Description: "desc_updated", Title: "title_updated",
		}

		notExisting := "not found"
		err = s.UpdateEvent(ctx, notExisting, &expEventU)
		require.ErrorIs(t, err, calendar.ErrEventNotFound)
		event, err := s.GetEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, &expEvent, event)
	})
}
