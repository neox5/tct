package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SenderMetrics holds all Prometheus metrics for sender mode.
type SenderMetrics struct {
	RequestsOk   prometheus.Counter
	RequestsErr  *prometheus.CounterVec
	ResponseTime prometheus.Histogram
	Inflight     prometheus.Gauge
}

// NewSenderMetrics creates and registers sender metrics with Prometheus.
func NewSenderMetrics() *SenderMetrics {
	return &SenderMetrics{
		RequestsOk: promauto.NewCounter(prometheus.CounterOpts{
			Name: "tct_sender_requests_ok_total",
			Help: "Total number of successful requests (HTTP 200)",
		}),

		RequestsErr: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tct_sender_requests_err_total",
				Help: "Total number of failed requests by error class",
			},
			[]string{"class"},
		),

		ResponseTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "tct_sender_response_time_seconds",
			Help: "HTTP request latency distribution",
			// Use default buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		}),

		Inflight: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "tct_sender_inflight",
			Help: "Number of currently in-flight requests",
		}),
	}
}

// RecordSuccess increments the success counter.
func (m *SenderMetrics) RecordSuccess() {
	m.RequestsOk.Inc()
}

// RecordError increments the error counter for the specified class.
// Valid classes: "timeout", "http_500", "conn", "other"
func (m *SenderMetrics) RecordError(class string) {
	m.RequestsErr.WithLabelValues(class).Inc()
}

// ObserveResponseTime records a request latency in seconds.
func (m *SenderMetrics) ObserveResponseTime(seconds float64) {
	m.ResponseTime.Observe(seconds)
}

// InflightInc increments the in-flight request counter.
// Call this before starting a request.
func (m *SenderMetrics) InflightInc() {
	m.Inflight.Inc()
}

// InflightDec decrements the in-flight request counter.
// Call this after request completes (success or failure).
func (m *SenderMetrics) InflightDec() {
	m.Inflight.Dec()
}
