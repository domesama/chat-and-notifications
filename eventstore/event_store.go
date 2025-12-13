package eventstore

import (
	"context"
	"time"

	"github.com/domesama/chat-and-notifications/event/eventmsg"
)

// EventStore is an interface for deduplication of events (e.g., based on timestamps). According to the usage, this project has no usage for it yet.
type EventStore[MsgValue any] interface {
	FilterInvalidMessage(ctx context.Context, msg eventmsg.Message[MsgValue]) (
		filteredMessage eventmsg.Message[MsgValue], shouldDropEntirely bool,
	)
	WriteEventStore(ctx context.Context, msg eventmsg.Message[MsgValue]) error
}

type NoOpEventStore[MsgValue any] struct {
}

func (n NoOpEventStore[MsgValue]) FilterInvalidMessage(ctx context.Context, msg eventmsg.Message[MsgValue]) (
	filteredMessages eventmsg.Message[MsgValue], shouldDropEntirely bool,
) {
	return msg, false
}

func (n NoOpEventStore[MsgValue]) WriteEventStore(ctx context.Context, msg eventmsg.Message[MsgValue]) error {
	return nil
}

type UpdatedAndPublishedTimeEventStore struct {
	DataUpdatedTime time.Time `json:"updated"`
	PublishedTime   time.Time `json:"published"`
}

func (e UpdatedAndPublishedTimeEventStore) IsIncomingTimeAfter(
	incomingUpdatedTime time.Time,
	incomingPublished time.Time,
) bool {
	if incomingUpdatedTime.After(e.DataUpdatedTime) {
		return true
	}
	if incomingUpdatedTime.Equal(e.DataUpdatedTime) && incomingPublished.After(e.PublishedTime) {
		return true
	}
	return false
}
