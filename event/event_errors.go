package event

import "errors"

var (
	ErrHandleMessageFailed          = errors.New("handle message failed")
	ErrEventUnableToWriteEventStore = errors.New("unable to write event store")
)
