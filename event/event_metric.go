package event

import (
	"context"
	"log/slog"

	"github.com/domesama/chat-and-notifications/event/eventmsg"
	doakesmetrics "github.com/domesama/doakes/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type EventMetric struct {
	Name               string
	SuccessEventMetric metric.Int64Counter
	RetryEventMetric   metric.Int64Counter
	DroppedEventMetric metric.Int64Counter
}

func (m EventMetric) IncrementDropDueToInvalidEvent(ctx context.Context) {
	slog.Warn("[HandleEvent] Event dropped due to", "reason", EventReasonInvalidEvent)
	m.DroppedEventMetric.Add(
		ctx, 1,
		CreateMetricLabel(EventAttributeDropReason, EventReasonInvalidEvent),
	)
}

func (m EventMetric) IncrementDropDueToFailedEventStoreValidation(ctx context.Context) {
	slog.Warn(
		"[HandleEvent] Event entirely dropped due to",
		"reason", EventReasonDroppedFromEventStoreValidation,
	)

	m.DroppedEventMetric.Add(
		ctx, 1,
		CreateMetricLabel(EventAttributeDropReason, EventReasonDroppedFromEventStoreValidation),
	)
}

func (m EventMetric) IncrementDropWithCustomReason(ctx context.Context, reason string) {
	slog.Warn("[HandleEvent] Event dropped due to", "reason", reason)
	m.DroppedEventMetric.Add(
		ctx, 1,
		CreateMetricLabel(EventAttributeDropReason, MetricLabelValue(reason)),
	)
}

func CreateEventMetrics(name string) *EventMetric {
	successCounter, err := doakesmetrics.GetDefaultMeter().Int64Counter(SuccessEventMetricType.GetMetricName(name))
	if err != nil {
		panic(err)
	}
	failedCounter, err := doakesmetrics.GetDefaultMeter().Int64Counter(FailedEventMetricType.GetMetricName(name))
	if err != nil {
		panic(err)
	}
	droppedCounter, err := doakesmetrics.GetDefaultMeter().Int64Counter(DroppedEventMetricType.GetMetricName(name))
	if err != nil {
		panic(err)
	}
	return &EventMetric{
		Name:               name,
		SuccessEventMetric: successCounter,
		RetryEventMetric:   failedCounter,
		DroppedEventMetric: droppedCounter,
	}
}

func CreateEventTypeLabel[MsgValue any](
	msgHandler BaseMessageHandler[MsgValue],
	message eventmsg.Message[MsgValue],
) metric.MeasurementOption {
	return metric.WithAttributes(
		attribute.String(EventAttributeEventType.ToString(), msgHandler.GetEventType(message)),
	)
}

func CreateMetricLabel(key MetricLabel, value MetricLabelValue) metric.MeasurementOption {
	return metric.WithAttributes(
		attribute.String(string(key), string(value)),
	)
}
