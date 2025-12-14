package event

import (
	"context"
	"log/slog"

	"github.com/IBM/sarama"
	"github.com/domesama/chat-and-notifications/event/eventmsg"
)

type BaseMessageHandler[MsgValue any] interface {
	GetEventType(msg eventmsg.Message[MsgValue]) string
	ConvertMessageValue(rawValues []byte) (value MsgValue, shouldDrop bool, err error)
}

type SingleMessageHandler[MsgValue any] interface {
	BaseMessageHandler[MsgValue]
	HandleMessage(ctx context.Context, messageValue eventmsg.Message[MsgValue]) error
}

type BatchMessageHandler[MsgValue any] interface {
	BaseMessageHandler[MsgValue]
	HandleMessages(ctx context.Context, eventType string, messageValue ...eventmsg.Message[MsgValue]) error
}

func covertSaramaMessagePayload[MsgValue any](handler BaseMessageHandler[MsgValue], msg *sarama.ConsumerMessage) (
	res eventmsg.Message[MsgValue], shouldDrop bool,
) {

	if msg == nil || msg.Key == nil {
		return res, true
	}

	res = eventmsg.Message[MsgValue]{
		Key:       string(msg.Key),
		Headers:   convertKvHeaders(msg.Headers),
		Timestamp: msg.Timestamp,
	}

	value, shouldDrop, err := handler.ConvertMessageValue(msg.Value)

	if err != nil {
		slog.Error("[covertSaramaMessagePayload] Failed to convert message", "key", msg.Key, "error", err.Error())
		shouldDrop = true
	}

	res.Value = value
	return
}

func convertKvHeaders(headers []*sarama.RecordHeader) map[string][]string {
	res := make(map[string][]string)
	for _, header := range headers {
		if header == nil || string(header.Key) == "" {
			continue
		}
		key := string(header.Key)
		res[key] = append(res[key], string(header.Value))
	}
	return res
}
