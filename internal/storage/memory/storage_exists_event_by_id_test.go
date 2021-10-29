package memorystorage_test

import (
	"context"
	"testing"
	"time"

	"github.com/seregproj/calendar/internal/storage"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestExistsEventByID(t *testing.T) {
	t.Run("test exists/not exists event", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		now := time.Now()
		after := time.Now().Add(time.Second * 10)

		// adding to empty map
		expEvent := storage.Event{ID: "test", Start: now, Finish: after, Description: "desc1", Title: "title1"}

		// not exists
		exists, err := s.ExistsEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, false, exists)

		err = s.CreateEvent(ctx, &expEvent)
		require.NoError(t, err)

		// exists
		exists, err = s.ExistsEventByID(ctx, expEvent.ID)
		require.NoError(t, err)
		require.Equal(t, true, exists)

		// not exists other item
		notExists := "test2"
		exists, err = s.ExistsEventByID(ctx, notExists)
		require.NoError(t, err)
		require.Equal(t, false, exists)
	})
}
