package event

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/domesama/chat-and-notifications/eventstore"
	"github.com/domesama/chat-and-notifications/utils"
)

type SingleEventHandler[MsgValue any] struct {
	MessageHandler SingleMessageHandler[MsgValue]
	EventMetric    *EventMetric
	EventStore     eventstore.EventStore[MsgValue]
}

func NewSingleEventHandler[MsgValue any](
	handler SingleMessageHandler[MsgValue],
	metric *EventMetric,
	options ...SingleEventHandlerOptions[MsgValue],
) SingleEventHandler[MsgValue] {

	optionalParam := bindEventHandlerOptions[MsgValue](options...)
	return SingleEventHandler[MsgValue]{
		MessageHandler: handler,
		EventMetric:    metric,
		EventStore:     optionalParam.EventStore,
	}
}

func (e SingleEventHandler[MsgValue]) HandleEvent(ctx context.Context, msg *sarama.ConsumerMessage) (err error) {
	message, shouldDrop := covertSaramaMessagePayload(e.MessageHandler, msg)

	if shouldDrop {
		e.EventMetric.IncrementDropDueToInvalidEvent(ctx)
		return nil
	}

	message, shouldDropEntirely := e.EventStore.FilterInvalidMessage(ctx, message)
	if shouldDropEntirely {
		e.EventMetric.IncrementDropDueToFailedEventStoreValidation(ctx)
		return nil
	}

	if err := e.MessageHandler.HandleMessage(ctx, message); err != nil {
		err = utils.WrapError(err, ErrHandleMessageFailed)

		slog.Error(err.Error())
		e.EventMetric.RetryEventMetric.Add(ctx, 1, CreateEventTypeLabel(e.MessageHandler, message))

		return err
	}
	if err = e.EventStore.WriteEventStore(ctx, message); err != nil {
		_ = utils.WrapError(
			fmt.Errorf("%w:%+v", err, message), ErrEventUnableToWriteEventStore,
		)

		return nil
	}

	e.EventMetric.SuccessEventMetric.Add(ctx, 1, CreateEventTypeLabel(e.MessageHandler, message))
	return nil
}
