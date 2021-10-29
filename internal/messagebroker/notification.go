package messagebroker

import "time"

type Notification struct {
	EventID    string
	EventTitle string
	EventStart time.Time
}

func NewNotification(eventID, eventTitle string, eventStart time.Time) *Notification {
	return &Notification{EventID: eventID, EventTitle: eventTitle, EventStart: eventStart}
}
