package tezos

import (
	"context"
	"testing"
	"time"
)

func TestRetryDelay(t *testing.T) {
	_, c, _, done := newTestConnector(t)
	defer done()

	c.retry.MaximumDelay = 1 * time.Microsecond
	c.retry.InitialDelay = 100 * time.Microsecond

	c.doFailureDelay(context.Background(), 1)
	c.doFailureDelay(context.Background(), 10)
}
