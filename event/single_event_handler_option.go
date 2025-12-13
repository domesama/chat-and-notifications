package event

import (
	"context"

	"github.com/domesama/chat-and-notifications/eventstore"
)

type SingleEventHandlerOptionalParams[MsgValue any] struct {
	EventStore         eventstore.EventStore[MsgValue]
	ShouldRetryMessage func(ctx context.Context, name string, msg MsgValue)
}

type SingleEventHandlerOptions[MsgValue any] func(optionalParam *SingleEventHandlerOptionalParams[MsgValue])

func WithEventStore[MsgValue any](eventStore eventstore.EventStore[MsgValue]) SingleEventHandlerOptions[MsgValue] {
	return func(optionalParam *SingleEventHandlerOptionalParams[MsgValue]) {
		optionalParam.EventStore = eventStore
	}
}

func bindEventHandlerOptions[MsgValue any](opts ...SingleEventHandlerOptions[MsgValue]) SingleEventHandlerOptionalParams[MsgValue] {
	optionalParam := SingleEventHandlerOptionalParams[MsgValue]{
		EventStore: eventstore.NoOpEventStore[MsgValue]{},
	}
	for _, opt := range opts {
		opt(&optionalParam)
	}
	return optionalParam
}
