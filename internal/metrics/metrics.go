// Package metrics provides Prometheus metric definitions and HTTP handler for tct.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Handler returns an HTTP handler for the /metrics endpoint.
// This handler exposes all registered Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}
