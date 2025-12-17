// Package handler provides HTTP handlers for tct endpoints.
package handler

import (
	"net/http"
)

// Healthz handles GET /healthz requests.
// Always returns 200 OK if the process is running.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// Readyz handles GET /readyz requests.
// Always returns 200 OK after initialization is complete.
// Unlike healthz, this is unaffected by simulated outages, hangs, or delays.
func Readyz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
