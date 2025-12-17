// Package handler provides HTTP handlers for tct endpoints.
package handler

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/neox5/tct/internal/config"
	"github.com/neox5/tct/internal/logger"
	"github.com/neox5/tct/internal/metrics"
)

// InboxHandler creates a handler for POST /inbox with behavior injection.
func InboxHandler(cfg *config.Config, log *logger.Logger, m *metrics.ReceiverMetrics) http.HandlerFunc {
	// Initialize outage state
	outage := &outageState{
		cfg:   cfg,
		log:   log,
		mutex: &sync.RWMutex{},
	}

	// Start outage management if configured
	if cfg.OutageAfter > 0 && cfg.OutageFor > 0 {
		go outage.manage()
	}

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Check if outage is active
		if outage.isActive() {
			m.RecordRequest("outage")
			m.SetOutageState(true)
			// Block indefinitely during outage (no response)
			select {}
		}
		m.SetOutageState(false)

		// 2. Apply hang decision
		if rand.Float64() < cfg.HangRate {
			m.RecordRequest("hang")
			log.Debug("request hanging", "path", r.URL.Path)
			// Block indefinitely (no response)
			select {}
		}

		// 3. Apply response delay + jitter
		delay := cfg.ResponseDelay
		if cfg.ResponseJitter > 0 {
			jitter := time.Duration(rand.Int63n(int64(cfg.ResponseJitter)))
			delay += jitter
		}
		if delay > 0 {
			time.Sleep(delay)
		}

		// 4. Return error or success
		if rand.Float64() < cfg.ErrorRate {
			m.RecordRequest("error")
			m.ObserveHandlerTime(time.Since(start).Seconds())
			log.Debug("returning error", "path", r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		}

		m.RecordRequest("ok")
		m.ObserveHandlerTime(time.Since(start).Seconds())
		log.Debug("request successful", "path", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

// outageState manages the outage lifecycle.
type outageState struct {
	cfg    *config.Config
	log    *logger.Logger
	active bool
	mutex  *sync.RWMutex
}

// isActive returns whether an outage is currently active.
func (o *outageState) isActive() bool {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.active
}

// setActive sets the outage state.
func (o *outageState) setActive(active bool) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	o.active = active
}

// manage runs the outage lifecycle loop.
func (o *outageState) manage() {
	// Wait for initial delay
	time.Sleep(o.cfg.OutageAfter)

	for {
		// Start outage
		o.log.Info("outage started", "duration", o.cfg.OutageFor)
		o.setActive(true)
		time.Sleep(o.cfg.OutageFor)

		// End outage
		o.log.Info("outage ended")
		o.setActive(false)

		// If not repeating, stop
		if !o.cfg.OutageRepeat {
			return
		}

		// Wait for next cycle
		time.Sleep(o.cfg.OutageAfter)
	}
}
