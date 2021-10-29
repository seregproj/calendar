package internalgrpc

import (
	"context"
	"errors"

	pb "github.com/seregproj/calendar/api/proto"
	"github.com/seregproj/calendar/internal/app/calendar"
	"github.com/seregproj/calendar/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EventServer struct {
	app Application
	pb.UnimplementedEventServiceServer
}

func NewEventServer(app Application) *EventServer {
	return &EventServer{app: app}
}

type Application interface {
	CreateEvent(ctx context.Context, event *storage.Event) (string, error)
	UpdateEvent(ctx context.Context, uuid string, event *storage.Event) error
	DeleteEvent(ctx context.Context, uuid string) error
	GetEventsByDay(ctx context.Context, date string, limit, offset int64) ([]*storage.Event, error)
}

func toAppEvent(re *pb.Event) (*storage.Event, error) {
	return storage.NewEvent("", re.GetTitle(), re.GetDescription(), re.GetDateStart().AsTime(),
		re.GetDateFinish().AsTime())
}

func fromAppEvent(event *storage.Event) *pb.Event {
	pbe := pb.Event{
		Title:       event.Title,
		Description: event.Description,
		DateStart:   timestamppb.New(event.Start),
		DateFinish:  timestamppb.New(event.Finish),
	}

	return &pbe
}

func (s EventServer) CreateEvent(ctx context.Context, req *pb.Event) (*pb.CreateEventResponse, error) {
	e, err := toAppEvent(req)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	id, err := s.app.CreateEvent(ctx, e)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.CreateEventResponse{Uuid: id}, nil
}

func (s EventServer) UpdateEvent(ctx context.Context, req *pb.UpdateEventRequest) (*pb.UpdateEventResponse, error) {
	e, err := toAppEvent(req.GetEvent())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	euuid := req.GetUuid()
	err = s.app.UpdateEvent(ctx, euuid, e)
	if err != nil {
		if errors.Is(err, calendar.ErrEventNotFound) {
			return nil, status.Errorf(codes.InvalidArgument, calendar.ErrEventNotFound.Error())
		}

		return nil, status.Errorf(codes.Internal, calendar.ErrUnexpected.Error())
	}

	return &pb.UpdateEventResponse{}, nil
}

func (s EventServer) DeleteEvent(ctx context.Context, req *pb.DeleteEventRequest) (*pb.DeleteEventResponse, error) {
	if err := s.app.DeleteEvent(ctx, req.GetUuid()); err != nil && !errors.Is(err, calendar.ErrEventNotFound) {
		return nil, status.Errorf(codes.Internal, calendar.ErrUnexpected.Error())
	}

	return &pb.DeleteEventResponse{}, nil
}

func (s EventServer) GetEventsByDay(ctx context.Context, req *pb.GetEventsByDayRequest) (*pb.Events, error) {
	events, err := s.app.GetEventsByDay(ctx, req.GetDay(), req.GetLimit(), req.GetOffset())
	if err != nil {
		if errors.Is(err, calendar.ErrInvalidDateFormat) {
			return nil, status.Errorf(codes.InvalidArgument, calendar.ErrInvalidDateFormat.Error())
		}

		return nil, status.Errorf(codes.Internal, calendar.ErrUnexpected.Error())
	}

	pbEvents := make([]*pb.Event, 0, len(events))
	for _, event := range events {
		pbEvents = append(pbEvents, fromAppEvent(event))
	}

	return &pb.Events{Items: pbEvents}, nil
}
