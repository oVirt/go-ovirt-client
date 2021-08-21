// This file contains tests for the internal retry functionality. It is therefore excluded from the testpackage check.

package ovirtclient // nolint:testpackage

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type retryFail struct {
	failCount int
}

func (r *retryFail) run() error {
	r.failCount++
	return fmt.Errorf("test failure")
}

func TestTimeoutStrategy(t *testing.T) {
	t.Parallel()
	r := &retryFail{}
	startTime := time.Now()
	err := retry(
		"test",
		nil,
		[]RetryStrategy{
			ExponentialBackoff(1),
			Timeout(3 * time.Second),
		},
		r.run,
	)
	endTime := time.Now()
	if err == nil {
		t.Fatalf("retry on a failing call did not return with an error")
	}
	// We allow for a little leeway due to the timing
	if r.failCount < 2 {
		t.Fatalf("retry didn't call the target function enough times (%d)", r.failCount)
	}
	if r.failCount > 4 {
		t.Fatalf("retry called too many times (%d)", r.failCount)
	}
	if elapsedTime := endTime.Sub(startTime).Seconds(); elapsedTime < 2 {
		t.Fatalf("retry didn't run for enough time (%f seconds)", elapsedTime)
	}
}

func TestContextStrategy(t *testing.T) {
	t.Parallel()

	var err error
	var startTime time.Time
	var endTime time.Time
	done := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	r := &retryFail{}
	go func() {
		startTime = time.Now()
		err = retry(
			"test",
			nil,
			[]RetryStrategy{
				ExponentialBackoff(1),
				ContextStrategy(ctx),
			},
			r.run,
		)
		endTime = time.Now()
		close(done)
	}()

	<-time.After(4 * time.Second)

	cancel()

	<-done

	if err == nil {
		t.Fatalf("retry on a failing call did not return with an error")
	}
	if r.failCount < 3 {
		t.Fatalf("retry didn't call the target function enough times (%d)", r.failCount)
	}
	// Add some leeway for slower machines.
	if r.failCount > 5 {
		t.Fatalf("retry called too many times (%d)", r.failCount)
	}
	if endTime.Sub(startTime).Seconds() < 3 {
		t.Fatalf("retry didn't run for enough time")
	}
}
