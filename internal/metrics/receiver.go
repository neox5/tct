package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ReceiverMetrics holds all Prometheus metrics for receiver mode.
type ReceiverMetrics struct {
	RequestsTotal *prometheus.CounterVec
	HandlerTime   prometheus.Histogram
	OutageState   prometheus.Gauge
}

// NewReceiverMetrics creates and registers receiver metrics with Prometheus.
func NewReceiverMetrics() *ReceiverMetrics {
	return &ReceiverMetrics{
		RequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tct_receiver_requests_total",
				Help: "Total number of received requests by outcome",
			},
			[]string{"outcome"},
		),

		HandlerTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "tct_receiver_handler_time_seconds",
			Help: "Handler execution time distribution",
			// Use default buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		}),

		OutageState: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "tct_receiver_outage_state",
			Help: "Current outage state (0=normal, 1=outage)",
		}),
	}
}

// RecordRequest increments the request counter for the specified outcome.
// Valid outcomes: "ok", "error", "hang", "outage"
func (m *ReceiverMetrics) RecordRequest(outcome string) {
	m.RequestsTotal.WithLabelValues(outcome).Inc()
}

// ObserveHandlerTime records handler execution time in seconds.
func (m *ReceiverMetrics) ObserveHandlerTime(seconds float64) {
	m.HandlerTime.Observe(seconds)
}

// SetOutageState sets the outage state gauge.
// Use 0 for normal operation, 1 for active outage.
func (m *ReceiverMetrics) SetOutageState(active bool) {
	if active {
		m.OutageState.Set(1)
	} else {
		m.OutageState.Set(0)
	}
}
