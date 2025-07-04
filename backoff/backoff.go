package backoff

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/acoshift/pgsql"
)

// Config contains common configuration for all backoff strategies
type Config struct {
	BaseDelay time.Duration // Base delay for backoff
	MaxDelay  time.Duration // Maximum delay cap
}

// ExponentialConfig contains configuration for exponential backoff
type ExponentialConfig struct {
	Config
	Multiplier float64 // Multiplier for exponential growth
	JitterType JitterType
}

// LinearConfig contains configuration for linear backoff
type LinearConfig struct {
	Config
	Increment time.Duration // Amount to increase delay each attempt
}

// JitterType defines the type of jitter to apply
type JitterType int

const (
	// NoJitter applies no jitter
	NoJitter JitterType = iota
	// FullJitter applies full jitter (0 to calculated delay)
	FullJitter
	// EqualJitter applies equal jitter (half fixed + half random)
	EqualJitter
)

// NewExponential creates a new exponential backoff function
func NewExponential(config ExponentialConfig) pgsql.BackoffDelayFunc {
	return func(attempt int) time.Duration {
		baseDelay := time.Duration(float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt)))
		if baseDelay > config.MaxDelay {
			baseDelay = config.MaxDelay
		}

		var delay time.Duration
		switch config.JitterType {
		case FullJitter:
			// Full jitter: random delay between 0 and calculated delay
			if baseDelay > 0 {
				delay = time.Duration(rand.Int64N(int64(baseDelay)))
			} else {
				delay = baseDelay
			}
		case EqualJitter:
			// Equal jitter: half fixed + half random
			half := baseDelay / 2
			if half > 0 {
				delay = half + time.Duration(rand.Int64N(int64(half)))
			} else {
				delay = baseDelay
			}
		default:
			delay = baseDelay
		}

		return delay
	}
}

// NewLinear creates a new linear backoff function
func NewLinear(config LinearConfig) pgsql.BackoffDelayFunc {
	return func(attempt int) time.Duration {
		delay := config.BaseDelay + time.Duration(attempt)*config.Increment
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
		return delay
	}
}

func DefaultExponential() pgsql.BackoffDelayFunc {
	return NewExponential(ExponentialConfig{
		Config: Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  5 * time.Second,
		},
		Multiplier: 2.0,
		JitterType: NoJitter,
	})
}

func DefaultExponentialWithFullJitter() pgsql.BackoffDelayFunc {
	return NewExponential(ExponentialConfig{
		Config: Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  5 * time.Second,
		},
		Multiplier: 2.0,
		JitterType: FullJitter,
	})
}

func DefaultExponentialWithEqualJitter() pgsql.BackoffDelayFunc {
	return NewExponential(ExponentialConfig{
		Config: Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  5 * time.Second,
		},
		Multiplier: 2.0,
		JitterType: EqualJitter,
	})
}

func DefaultLinear() pgsql.BackoffDelayFunc {
	return NewLinear(LinearConfig{
		Config: Config{
			BaseDelay: 100 * time.Millisecond,
			MaxDelay:  5 * time.Second,
		},
		Increment: 100 * time.Millisecond,
	})
}
