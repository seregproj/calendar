package memorystorage_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/seregproj/calendar/internal/app/calendar"
	"github.com/seregproj/calendar/internal/storage"
	memorystorage "github.com/seregproj/calendar/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestUpdateEventAsProcessed(t *testing.T) {
	t.Run("update unexisting event", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		event := storage.Event{
			ID: "test", Start: time.Now(), Finish: time.Now().Add(time.Minute), Description: "desc1",
			Title: "title1",
		}
		err := s.UpdateEventAsProcessed(ctx, &event)
		require.ErrorIs(t, err, calendar.ErrEventNotFound)
	})

	t.Run("update existing event", func(t *testing.T) {
		ctx := context.Background()
		s := memorystorage.New()

		event1 := storage.Event{
			ID: "event1", Start: time.Now(), Finish: time.Now().Add(time.Minute),
			Description: "desc1", Title: "title1",
		}
		err := s.CreateEvent(ctx, &event1)
		require.NoError(t, err)

		eventForUpdate := storage.Event{
			ID: "event2", Start: time.Now(), Finish: time.Now().Add(time.Minute),
			Description: "desc1", Title: "title1",
		}
		err = s.CreateEvent(ctx, &eventForUpdate)
		require.NoError(t, err)

		event3 := storage.Event{
			ID: "event3", Start: time.Now(), Finish: time.Now().Add(time.Minute),
			Description: "desc1", Title: "title1",
		}
		err = s.CreateEvent(ctx, &event3)
		require.NoError(t, err)

		err = s.UpdateEventAsProcessed(ctx, &eventForUpdate)
		require.NoError(t, err)

		for _, v := range reflect.ValueOf(s).Elem().FieldByName("events").MapKeys() {
			var exp bool

			if v.String() == eventForUpdate.ID {
				exp = true
			}

			require.Equal(t, exp, reflect.ValueOf(s).Elem().FieldByName("events").MapIndex(v).Elem().
				FieldByName("Processed").Bool())
		}

		// repeatable update
		err = s.UpdateEventAsProcessed(ctx, &eventForUpdate)
		require.NoError(t, err)

		for _, v := range reflect.ValueOf(s).Elem().FieldByName("events").MapKeys() {
			var exp bool

			if v.String() == eventForUpdate.ID {
				exp = true
			}

			require.Equal(t, exp, reflect.ValueOf(s).Elem().FieldByName("events").MapIndex(v).Elem().
				FieldByName("Processed").Bool())
		}

		// update other event doesnt affect to initial state
		err = s.UpdateEventAsProcessed(ctx, &event1)
		require.NoError(t, err)

		for _, v := range reflect.ValueOf(s).Elem().FieldByName("events").MapKeys() {
			var exp bool

			if v.String() == event1.ID || v.String() == eventForUpdate.ID {
				exp = true
			}

			require.Equal(t, exp, reflect.ValueOf(s).Elem().FieldByName("events").MapIndex(v).Elem().
				FieldByName("Processed").Bool())
		}
	})
}
