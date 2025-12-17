// Package config defines configuration structures for tct.
// Configuration is loaded from environment variables via the env package.
package config

import "time"

// Config holds the complete application configuration.
// All fields are at the top level. The Mode field determines which
// subset of fields are relevant for the current execution.
type Config struct {
	// Common fields
	Mode     string `env:"TCT_MODE,required"`
	LogLevel string `env:"TCT_LOG_LEVEL,default=info"`

	// Sender fields
	SenderPort     int           `env:"TCT_SENDER_PORT,default=9090,min=1,max=65535"`
	ReceiverHost   string        `env:"TCT_RECEIVER_HOST,default=localhost"`
	ReceiverPort   int           `env:"TCT_RECEIVER_PORT,default=8080,min=1,max=65535"`
	RPS            float64       `env:"TCT_RPS,default=1.0,min=0"`
	StartDelay     time.Duration `env:"TCT_START_DELAY,default=0s"`
	RequestTimeout time.Duration `env:"TCT_REQUEST_TIMEOUT,default=2s,min=0s"`

	// Receiver fields
	ResponseDelay  time.Duration `env:"TCT_RESPONSE_DELAY,default=0s,min=0s"`
	ResponseJitter time.Duration `env:"TCT_RESPONSE_JITTER,default=0s,min=0s"`
	HangRate       float64       `env:"TCT_HANG_RATE,default=0,min=0,max=1"`
	ErrorRate      float64       `env:"TCT_ERROR_RATE,default=0,min=0,max=1"`
	OutageAfter    time.Duration `env:"TCT_OUTAGE_AFTER,default=0s,min=0s"`
	OutageFor      time.Duration `env:"TCT_OUTAGE_FOR,default=0s,min=0s"`
	OutageRepeat   bool          `env:"TCT_OUTAGE_REPEAT,default=false"`
}
