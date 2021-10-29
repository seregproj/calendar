//go:build integration
// +build integration

package integration_test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/seregproj/calendar/api/proto"
	"github.com/seregproj/calendar/internal/app/calendar"
	"github.com/seregproj/calendar/internal/storage"
	sqlstorage "github.com/seregproj/calendar/internal/storage/sql"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventsSuite struct {
	suite.Suite
	db          *pgxpool.Pool
	eventClient proto.EventServiceClient
	ctx         context.Context
}

func (s *EventsSuite) SetupSuite() {
	s.ctx = context.Background()
	grpcConn, err := grpc.Dial(net.JoinHostPort(os.Getenv("GRPC_HOST"), os.Getenv("GRPC_PORT")), grpc.WithInsecure())
	s.Require().NoError(err)

	s.eventClient = proto.NewEventServiceClient(grpcConn)

	connPool, err := pgxpool.Connect(s.ctx, os.Getenv("PGSQL_DSN"))
	s.Require().NoError(err)
	s.db = connPool
}

func (s *EventsSuite) TearDownTest() {
	_, err := s.db.Exec(s.ctx, "TRUNCATE events")
	s.Require().NoError(err)
}

func (s *EventsSuite) TestCreateEventSimple() {
	event1 := getRandEvent(time.Now().Add(time.Minute*2), time.Now().Add(time.Minute*5))
	event2 := getRandEvent(time.Now().Add(time.Minute*3), time.Now().Add(time.Minute*7))

	resp1, err := s.eventClient.CreateEvent(s.ctx, event1)
	s.Require().NoError(err)
	s.Require().Len(resp1.GetUuid(), 36)

	resp2, err := s.eventClient.CreateEvent(s.ctx, event2)
	s.Require().NoError(err)
	s.Require().Len(resp2.GetUuid(), 36)

	event1Db := s.getEvent(resp1.GetUuid())
	s.Require().Equal(event1.GetTitle(), event1Db.Title)
	s.Require().Equal(event1.GetDescription(), event1Db.Description)
	ds, err := time.Parse("2006-01-02 15:04", event1.GetDateStart().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(ds, event1Db.Start)
	df, err := time.Parse("2006-01-02 15:04", event1.GetDateFinish().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(df, event1Db.Finish)

	event2Db := s.getEvent(resp2.GetUuid())
	s.Require().Equal(event2.GetTitle(), event2Db.Title)
	s.Require().Equal(event2.GetDescription(), event2Db.Description)
	ds, err = time.Parse("2006-01-02 15:04", event2.GetDateStart().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(ds, event2Db.Start)
	df, err = time.Parse("2006-01-02 15:04", event2.GetDateFinish().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(df, event2Db.Finish)
}

func (s *EventsSuite) TestCreateEventStartBeforeNow() {
	event := getRandEvent(time.Now().Add(-time.Minute*2), time.Now().Add(time.Minute*5))

	_, err := s.eventClient.CreateEvent(s.ctx, event)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
	s.Require().Equal(st.Message(), fmt.Sprintf("invalid datestart: %v, datestart should be in future",
		event.DateStart.AsTime()))
}

func (s *EventsSuite) TestCreateEventStartAfterFinish() {
	event := getRandEvent(time.Now().Add(time.Minute*5), time.Now().Add(time.Minute*2))

	_, err := s.eventClient.CreateEvent(s.ctx, event)
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
	s.Require().Equal(st.Message(), fmt.Sprintf("invalid datestart: %v, datestart should be before datefinish",
		event.DateStart.AsTime()))
}

func (s *EventsSuite) TestUpdateEventSimple() {
	event := getRandEvent(time.Now().Add(time.Minute*2), time.Now().Add(time.Minute*5))

	resp, err := s.eventClient.CreateEvent(s.ctx, event)
	s.Require().NoError(err)

	_, err = s.eventClient.UpdateEvent(s.ctx, &proto.UpdateEventRequest{Uuid: resp.GetUuid(), Event: event})
	s.Require().NoError(err)

	eventDB := s.getEvent(resp.GetUuid())
	s.Require().Equal(event.GetTitle(), eventDB.Title)
	s.Require().Equal(event.GetDescription(), eventDB.Description)
	ds, err := time.Parse("2006-01-02 15:04", event.GetDateStart().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(ds, eventDB.Start)
	df, err := time.Parse("2006-01-02 15:04", event.GetDateFinish().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(df, eventDB.Finish)
}

func (s *EventsSuite) TestUpdateUnexistingEvent() {
	event := getRandEvent(time.Now().Add(time.Minute*2), time.Now().Add(time.Minute*5))

	_, err := s.eventClient.CreateEvent(s.ctx, event)
	s.Require().NoError(err)

	eventID, err := uuid.NewV4()
	s.Require().NotNil(eventID)
	s.Require().NoError(err)
	_, err = s.eventClient.UpdateEvent(s.ctx, &proto.UpdateEventRequest{Uuid: eventID.String(), Event: event})
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
	s.Require().Equal(calendar.ErrEventNotFound.Error(), st.Message())
}

func (s *EventsSuite) TestUpdateEventStartBeforeNow() {
	event := getRandEvent(time.Now().Add(time.Minute*2), time.Now().Add(time.Minute*5))

	resp, err := s.eventClient.CreateEvent(s.ctx, event)
	s.Require().NoError(err)

	event.DateStart = timestamppb.New(time.Now().Add(-time.Minute * 2))
	_, err = s.eventClient.UpdateEvent(s.ctx, &proto.UpdateEventRequest{Uuid: resp.GetUuid(), Event: event})
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
	s.Require().Equal(st.Message(), fmt.Sprintf("invalid datestart: %v, datestart should be in future",
		event.DateStart.AsTime()))
}

func (s *EventsSuite) TestUpdateEventStartAfterFinish() {
	event := getRandEvent(time.Now().Add(time.Minute*4), time.Now().Add(time.Minute*5))

	resp, err := s.eventClient.CreateEvent(s.ctx, event)
	s.Require().NoError(err)

	event.DateFinish = timestamppb.New(time.Now().Add(time.Minute * 2))
	_, err = s.eventClient.UpdateEvent(s.ctx, &proto.UpdateEventRequest{Uuid: resp.GetUuid(), Event: event})
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.InvalidArgument, st.Code())
	s.Require().Equal(st.Message(), fmt.Sprintf("invalid datestart: %v, datestart should be before datefinish",
		event.DateStart.AsTime()))
}

func (s *EventsSuite) TestDeleteEventSimple() {
	event1 := getRandEvent(time.Now().Add(time.Minute*2), time.Now().Add(time.Minute*5))
	resp1, err := s.eventClient.CreateEvent(s.ctx, event1)
	s.Require().NoError(err)

	event2 := getRandEvent(time.Now().Add(time.Minute*3), time.Now().Add(time.Minute*4))
	resp2, err := s.eventClient.CreateEvent(s.ctx, event2)
	s.Require().NoError(err)

	_, err = s.eventClient.DeleteEvent(s.ctx, &proto.DeleteEventRequest{Uuid: resp2.GetUuid()})
	s.Require().NoError(err)

	// check event1 aren't deleted
	eventDB := s.getEvent(resp1.GetUuid())
	s.Require().Equal(event1.GetTitle(), eventDB.Title)
	s.Require().Equal(event1.GetDescription(), eventDB.Description)
	ds, err := time.Parse("2006-01-02 15:04", event1.GetDateStart().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(ds, eventDB.Start)
	df, err := time.Parse("2006-01-02 15:04", event1.GetDateFinish().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(df, eventDB.Finish)

	// check event2 are deleted
	s.checkNoEvent(resp2.GetUuid())
}

func (s *EventsSuite) TestDeleteUnexistingEvent() {
	event := getRandEvent(time.Now().Add(time.Minute*2), time.Now().Add(time.Minute*5))
	resp, err := s.eventClient.CreateEvent(s.ctx, event)
	s.Require().NoError(err)

	eventID, err := uuid.NewV4()
	s.Require().NotNil(eventID)
	s.Require().NoError(err)
	_, err = s.eventClient.DeleteEvent(s.ctx, &proto.DeleteEventRequest{Uuid: eventID.String()})
	s.Require().NoError(err)

	// check event aren't deleted
	eventDB := s.getEvent(resp.GetUuid())
	s.Require().Equal(event.GetTitle(), eventDB.Title)
	s.Require().Equal(event.GetDescription(), eventDB.Description)
	ds, err := time.Parse("2006-01-02 15:04", event.GetDateStart().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(ds, eventDB.Start)
	df, err := time.Parse("2006-01-02 15:04", event.GetDateFinish().AsTime().Format("2006-01-02 15:04"))
	s.Require().NoError(err)
	s.Require().Equal(df, eventDB.Finish)
}

func (s *EventsSuite) TestGetEventsByDayNoEvents() {
	resp, err := s.eventClient.GetEventsByDay(s.ctx, &proto.GetEventsByDayRequest{Day: "2021-05-03", Limit: 10})
	s.Require().NoError(err)
	s.Require().Equal(0, len(resp.GetItems()))
}

func (s *EventsSuite) TestGetEventsByDaySimple() {
	dayFrom := time.Now().AddDate(0, 0, 15).Format("2006-01-02")
	dayFromTime, err := time.Parse("2006-01-02", dayFrom)
	s.Require().NoError(err)

	eventCorrect1 := getRandEvent(dayFromTime, dayFromTime.Add(time.Hour*5))
	_, err = s.eventClient.CreateEvent(s.ctx, eventCorrect1)
	s.Require().NoError(err)

	eventNextDay := getRandEvent(dayFromTime.AddDate(0, 0, 1), dayFromTime.AddDate(0, 0, 2))
	_, err = s.eventClient.CreateEvent(s.ctx, eventNextDay)
	s.Require().NoError(err)

	eventDayBefore := getRandEvent(dayFromTime.AddDate(0, 0, -1), dayFromTime.AddDate(0, 0, 2))
	_, err = s.eventClient.CreateEvent(s.ctx, eventDayBefore)
	s.Require().NoError(err)

	eventCorrect2 := getRandEvent(dayFromTime, dayFromTime.AddDate(0, 0, 5))
	_, err = s.eventClient.CreateEvent(s.ctx, eventCorrect2)
	s.Require().NoError(err)

	eventCorrect3 := getRandEvent(dayFromTime, dayFromTime.AddDate(0, 2, 0))
	_, err = s.eventClient.CreateEvent(s.ctx, eventCorrect3)
	s.Require().NoError(err)

	// limit = 2, offset 0, check first and second correct events
	resp, err := s.eventClient.GetEventsByDay(s.ctx, &proto.GetEventsByDayRequest{Day: dayFrom, Limit: 2})
	s.Require().NoError(err)
	s.Require().Equal(2, len(resp.GetItems()))
	s.Require().True(eventPbExists(resp, eventCorrect1))
	s.Require().True(eventPbExists(resp, eventCorrect2))

	// offset = 2, check third correct event
	resp, err = s.eventClient.GetEventsByDay(s.ctx, &proto.GetEventsByDayRequest{Day: dayFrom, Limit: 10, Offset: 2})
	s.Require().NoError(err)
	s.Require().Equal(1, len(resp.GetItems()))
	s.Require().True(eventPbExists(resp, eventCorrect3))
}

func (s *EventsSuite) getEvent(uuid string) storage.Event {
	var event sqlstorage.Event

	err := pgxscan.Get(s.ctx, s.db, &event,
		`SELECT id, title, description, datetime_start, datetime_finish FROM events WHERE id = $1`, uuid)
	s.Require().NoError(err)

	return event.ToApp()
}

func (s *EventsSuite) checkNoEvent(uuid string) {
	res, err := s.db.Exec(s.ctx, `SELECT 1 FROM events where id = $1`, uuid)
	s.Require().NoError(err)
	s.Require().Equal(0, int(res.RowsAffected()))
}

func eventPbExists(events *proto.Events, event *proto.Event) bool {
	for _, e := range events.GetItems() {
		if e.GetTitle() == event.GetTitle() &&
			e.GetDescription() == event.GetDescription() &&
			e.GetDateStart().AsTime() == event.GetDateStart().AsTime() &&
			e.GetDateFinish().AsTime() == event.GetDateFinish().AsTime() {
			return true
		}
	}

	return false
}

func getRandEvent(dateStart, dateFinish time.Time) *proto.Event {
	rand.Seed(time.Now().UnixNano())

	return &proto.Event{
		Title:       faker.Word(),
		Description: faker.Word(),
		DateStart:   timestamppb.New(dateStart),
		DateFinish:  timestamppb.New(dateFinish),
	}
}

func TestEventsSuite(t *testing.T) {
	suite.Run(t, new(EventsSuite))
}
