package event

import "fmt"

type (
	MetricType       string
	MetricLabel      string
	MetricLabelValue string
)

const (
	SuccessEventMetricType MetricType = "success_event"
	FailedEventMetricType  MetricType = "failed_event"
	DroppedEventMetricType MetricType = "dropped_event"
)

const (
	EventAttributeDropReason MetricLabel = "dropped_reason"
	EventAttributeEventType  MetricLabel = "event_type"
)

const (
	EventReasonInvalidEvent                    MetricLabelValue = "invalid_format"
	EventReasonDroppedFromEventStoreValidation MetricLabelValue = "dropped_from_event_store_validation"
)

func (m MetricType) GetMetricName(name string) string {
	return fmt.Sprintf("%v_%v", name, string(m))
}

func (m MetricType) GetMetricTotalName() string {
	return fmt.Sprintf("%v_total", string(m))
}

func (m MetricLabel) ToString() string {
	return string(m)
}

func (m MetricLabelValue) ToString() string {
	return string(m)
}
