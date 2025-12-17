// Package generator provides request generation logic for sender mode.
package generator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/neox5/tct/internal/config"
	"github.com/neox5/tct/internal/logger"
	"github.com/neox5/tct/internal/metrics"
)

// Run executes the sender request generation loop.
// It generates HTTP POST requests at the configured rate until the context is cancelled.
func Run(ctx context.Context, cfg *config.Config, log *logger.Logger, m *metrics.SenderMetrics) error {
	// Wait for start delay
	if cfg.StartDelay > 0 {
		log.Info("waiting before starting", "delay", cfg.StartDelay)
		select {
		case <-time.After(cfg.StartDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Create HTTP client
	client := &http.Client{
		Timeout: cfg.RequestTimeout,
	}

	// Calculate interval between requests
	interval := time.Duration(float64(time.Second) / cfg.RPS)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	target := fmt.Sprintf("http://%s:%d/inbox", cfg.ReceiverHost, cfg.ReceiverPort)
	log.Info("starting request generation", "target", target, "rps", cfg.RPS)

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping request generation")
			return ctx.Err()

		case <-ticker.C:
			go sendRequest(ctx, client, target, log, m)
		}
	}
}

// sendRequest sends a single HTTP POST request and records metrics.
func sendRequest(ctx context.Context, client *http.Client, target string, log *logger.Logger, m *metrics.SenderMetrics) {
	m.InflightInc()
	defer m.InflightDec()

	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	if err != nil {
		m.RecordError("other")
		log.Error("failed to create request", "error", err)
		return
	}

	resp, err := client.Do(req)
	duration := time.Since(start).Seconds()
	m.ObserveResponseTime(duration)

	if err != nil {
		// Classify error
		if ctx.Err() != nil {
			m.RecordError("timeout")
			log.Debug("request timeout", "target", target)
		} else {
			m.RecordError("conn")
			log.Debug("connection error", "target", target, "error", err)
		}
		return
	}
	defer resp.Body.Close()

	// Drain response body
	io.Copy(io.Discard, resp.Body)

	// Classify response
	switch resp.StatusCode {
	case http.StatusOK:
		m.RecordSuccess()
		log.Debug("request successful", "target", target, "duration", duration)

	case http.StatusInternalServerError:
		m.RecordError("http_500")
		log.Debug("request failed", "target", target, "status", resp.StatusCode)

	default:
		m.RecordError("other")
		log.Debug("unexpected status", "target", target, "status", resp.StatusCode)
	}
}
