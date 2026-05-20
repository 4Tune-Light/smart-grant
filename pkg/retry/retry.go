package retry

import (
	"context"
	"math"
	"time"
)

const (
	defaultAttempts = 3
	baseDelay       = 100 * time.Millisecond
	maxDelay        = 5 * time.Second
)

type Config struct {
	Attempts int
	Base     time.Duration
	MaxDelay time.Duration
}

func Do(ctx context.Context, fn func() error) error {
	return DoWithConfig(ctx, fn, Config{
		Attempts: defaultAttempts,
		Base:     baseDelay,
		MaxDelay: maxDelay,
	})
}

func DoWithConfig(ctx context.Context, fn func() error, cfg Config) error {
	if cfg.Attempts < 1 {
		cfg.Attempts = defaultAttempts
	}
	if cfg.Base <= 0 {
		cfg.Base = baseDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = maxDelay
	}

	var lastErr error
	for i := 0; i < cfg.Attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := fn(); err != nil {
			lastErr = err
			if i == cfg.Attempts-1 {
				break
			}
			delay := time.Duration(math.Min(
				float64(cfg.Base)*math.Pow(2, float64(i)),
				float64(cfg.MaxDelay),
			))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			continue
		}
		return nil
	}
	return lastErr
}
