package ittesthelper

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/domesama/chat-and-notifications/event"
	testutils2 "github.com/domesama/chat-and-notifications/utils/testutils"
	"github.com/domesama/doakes/server"
	doakestest "github.com/domesama/doakes/testutil"
	"github.com/stretchr/testify/assert"
)

type EventMetricHelper struct {
	t                *testing.T
	mu               sync.RWMutex
	internalPrefix   string
	eventMetric      *event.EventMetric
	PrometheusHelper *doakestest.PrometheusHelper
}

func NewMetricHelper(
	t *testing.T,
	telemetryServer *server.TelemetryServer,
	metric *event.EventMetric) EventMetricHelper {
	return EventMetricHelper{
		t: t, eventMetric: metric,
		PrometheusHelper: doakestest.NewPrometheusHelper(telemetryServer.GetRunningPort()),
	}
}

type Label struct {
	LabelName     string
	LabelValue    string
	ExpectedValue int
}

type metricAssertion map[string][]Label

func (m *EventMetricHelper) ReadAndAssertIncomingCounterMetrics(
	incoming metricAssertion,
	shouldAssert bool,
) (totalMetricCount int) {
	m.mu.RLock()
	prefix := m.internalPrefix
	m.mu.RUnlock()

	for metricToAssert, properties := range incoming {
		for _, property := range properties {

			metricLabelToAssert := make(map[string]string)
			if property.LabelName != "" {
				metricLabelToAssert[property.LabelName] = property.LabelValue
			}

			actualMetric := m.PrometheusHelper.ParseMetrics(m.t).GetSingle(
				m.t, prefix+"_"+metricToAssert,
				map[string]string{
					property.LabelName: property.LabelValue,
				},
			)
			if actualMetric != nil {
				actualValue := int(actualMetric.GetCounter().GetValue())
				if shouldAssert {
					testutils2.AssertEqualWithMessage(
						m.t,
						fmt.Sprintf(
							"eventMetric of %v_%v_%v_%v",
							m.internalPrefix, metricToAssert, property.LabelName, property.LabelValue,
						),
						property.ExpectedValue,
						actualValue,
					)
					continue
				}

				totalMetricCount += actualValue

			}
		}
	}
	return
}

func (m *EventMetricHelper) EventuallyAssertSelectedCounterMetrics(
	metricToAssert map[string][]Label,
	expectedTotalMetricCount int,
) {
	assert.Eventually(
		m.t, func() bool {
			actualTotalMetricsCount := m.ReadAndAssertIncomingCounterMetrics(metricToAssert, false)
			if actualTotalMetricsCount == expectedTotalMetricCount {
				m.ReadAndAssertIncomingCounterMetrics(metricToAssert, true)
				return true
			}
			return false
		}, time.Second*5, time.Millisecond*100, "metrics do not match expected within 20 seconds",
	)
}

func (m *EventMetricHelper) ResetEventMetric() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.internalPrefix = fmt.Sprintf("test_%d_%d", time.Now().UnixNano(), os.Getpid())
	*m.eventMetric = *event.CreateEventMetrics(m.internalPrefix)
}

func (m *EventMetricHelper) NeverExceedsMetricCount(
	metricToAssert map[string][]Label,
	expectedMaxCount int,
	maxWaitingTime time.Duration,
) {
	assert.Never(
		m.t, func() bool {
			actualCount := m.ReadAndAssertIncomingCounterMetrics(metricToAssert, false)
			return actualCount > expectedMaxCount
		}, maxWaitingTime, time.Millisecond*100,
		fmt.Sprintf("metric count exceeded %d within %v", expectedMaxCount, maxWaitingTime),
	)
}
