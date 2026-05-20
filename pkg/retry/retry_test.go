package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDo_Success(t *testing.T) {
	count := 0
	err := Do(context.Background(), func() error {
		count++
		return nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDo_RetryThenSuccess(t *testing.T) {
	count := 0
	err := DoWithConfig(context.Background(), func() error {
		count++
		if count < 3 {
			return errors.New("transient error")
		}
		return nil
	}, Config{Attempts: 5, Base: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond})
	assert.NoError(t, err)
	assert.Equal(t, 3, count)
}

func TestDo_MaxRetriesExceeded(t *testing.T) {
	count := 0
	err := DoWithConfig(context.Background(), func() error {
		count++
		return errors.New("persistent error")
	}, Config{Attempts: 3, Base: 10 * time.Millisecond, MaxDelay: 50 * time.Millisecond})
	assert.Error(t, err)
	assert.Equal(t, 3, count)
}

func TestDo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(ctx, func() error {
		return errors.New("should not retry")
	})
	assert.ErrorIs(t, err, context.Canceled)
}

func TestDo_ZeroAttemptsDefaults(t *testing.T) {
	count := 0
	err := DoWithConfig(context.Background(), func() error {
		count++
		return errors.New("err")
	}, Config{Attempts: 0, Base: 0, MaxDelay: 0})
	assert.Error(t, err)
	assert.Equal(t, 3, count)
}
