// Package env provides parsing of environment variables into Go structs
// using struct field tags.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// tagOptions holds parsed tag configuration.
type tagOptions struct {
	required   bool
	defaultVal string
	min        string
	max        string
}

// Parse loads configuration from environment variables into the provided struct.
// The struct must be passed as a pointer.
//
// Supported tags:
//   - required: Field must have a value set in environment
//   - default=<value>: Default value if environment variable not set
//   - min=<value>: Minimum allowed value (numeric types and durations)
//   - max=<value>: Maximum allowed value (numeric types and durations)
//
// Example:
//
//	type Config struct {
//	    Port int `env:"PORT,default=8080,min=1,max=65535"`
//	}
func Parse(cfg any) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("config must be a non-nil pointer")
	}

	return parseStruct(v.Elem())
}

// parseStruct recursively parses struct fields.
func parseStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Handle embedded structs (e.g., CommonConfig)
		if field.Anonymous {
			if err := parseStruct(fieldVal); err != nil {
				return err
			}
			continue
		}

		// Skip fields without env tag
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		// Parse tag options
		envKey, opts := parseTag(tag)

		// Get value from environment
		envVal, exists := os.LookupEnv(envKey)

		// Handle required/default
		if !exists {
			if opts.required {
				return fmt.Errorf("%s is required", envKey)
			}
			if opts.defaultVal != "" {
				envVal = opts.defaultVal
			} else {
				continue // Skip unset optional fields without defaults
			}
		}

		// Parse and set field value
		if err := setField(fieldVal, envVal, envKey); err != nil {
			return err
		}

		// Validate constraints
		if err := validateField(fieldVal, opts, envKey); err != nil {
			return err
		}
	}

	return nil
}

// parseTag parses an env tag string into key and options.
// Format: "ENV_KEY,option1,option2=value"
func parseTag(tag string) (envKey string, opts tagOptions) {
	parts := strings.Split(tag, ",")
	envKey = parts[0]

	for _, part := range parts[1:] {
		switch {
		case part == "required":
			opts.required = true
		case strings.HasPrefix(part, "default="):
			opts.defaultVal = strings.TrimPrefix(part, "default=")
		case strings.HasPrefix(part, "min="):
			opts.min = strings.TrimPrefix(part, "min=")
		case strings.HasPrefix(part, "max="):
			opts.max = strings.TrimPrefix(part, "max=")
		}
	}

	return envKey, opts
}

// setField converts the string value to the appropriate type and sets the field.
func setField(field reflect.Value, value string, envKey string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int64:
		// Handle time.Duration specially
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("%s: invalid duration %q: %w", envKey, value, err)
			}
			field.SetInt(int64(d))
		} else {
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("%s: invalid integer %q: %w", envKey, value, err)
			}
			field.SetInt(i)
		}

	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("%s: invalid float %q: %w", envKey, value, err)
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("%s: invalid boolean %q: %w", envKey, value, err)
		}
		field.SetBool(b)

	default:
		return fmt.Errorf("%s: unsupported type %v", envKey, field.Type())
	}

	return nil
}

// validateField validates field value against min/max constraints.
func validateField(field reflect.Value, opts tagOptions, envKey string) error {
	// No constraints to validate
	if opts.min == "" && opts.max == "" {
		return nil
	}

	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		// Handle time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d := time.Duration(field.Int())

			if opts.min != "" {
				minDur, err := time.ParseDuration(opts.min)
				if err != nil {
					return fmt.Errorf("%s: invalid min duration %q", envKey, opts.min)
				}
				if d < minDur {
					return fmt.Errorf("%s: must be >= %v, got %v", envKey, minDur, d)
				}
			}

			if opts.max != "" {
				maxDur, err := time.ParseDuration(opts.max)
				if err != nil {
					return fmt.Errorf("%s: invalid max duration %q", envKey, opts.max)
				}
				if d > maxDur {
					return fmt.Errorf("%s: must be <= %v, got %v", envKey, maxDur, d)
				}
			}
		} else {
			// Regular integer
			i := field.Int()

			if opts.min != "" {
				minVal, err := strconv.ParseInt(opts.min, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: invalid min value %q", envKey, opts.min)
				}
				if i < minVal {
					return fmt.Errorf("%s: must be >= %v, got %v", envKey, minVal, i)
				}
			}

			if opts.max != "" {
				maxVal, err := strconv.ParseInt(opts.max, 10, 64)
				if err != nil {
					return fmt.Errorf("%s: invalid max value %q", envKey, opts.max)
				}
				if i > maxVal {
					return fmt.Errorf("%s: must be <= %v, got %v", envKey, maxVal, i)
				}
			}
		}

	case reflect.Float64:
		f := field.Float()

		if opts.min != "" {
			minVal, err := strconv.ParseFloat(opts.min, 64)
			if err != nil {
				return fmt.Errorf("%s: invalid min value %q", envKey, opts.min)
			}
			if f < minVal {
				return fmt.Errorf("%s: must be >= %v, got %v", envKey, minVal, f)
			}
		}

		if opts.max != "" {
			maxVal, err := strconv.ParseFloat(opts.max, 64)
			if err != nil {
				return fmt.Errorf("%s: invalid max value %q", envKey, opts.max)
			}
			if f > maxVal {
				return fmt.Errorf("%s: must be <= %v, got %v", envKey, maxVal, f)
			}
		}
	}

	return nil
}
