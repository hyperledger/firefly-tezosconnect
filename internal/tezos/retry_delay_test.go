package tezos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryDelay(t *testing.T) {
	_, c, _, done := newTestConnector(t)
	defer done()

	c.retry.MaximumDelay = 100 * time.Microsecond
	c.retry.InitialDelay = 1 * time.Microsecond

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		cancel()
	}()

	testCases := []struct {
		name         string
		failureCount int
		ctx          context.Context
		result       bool
	}{
		{
			name:         "zero failure count",
			failureCount: 0,
			ctx:          context.Background(),
			result:       false,
		},
		{
			name:         "retry delay exceeds max delay",
			failureCount: 10,
			ctx:          context.Background(),
			result:       false,
		},
		{
			name:         "ctx done",
			failureCount: 10,
			ctx:          ctx,
			result:       true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := c.doFailureDelay(tc.ctx, tc.failureCount)
			assert.Equal(t, tc.result, res)
		})
	}
}
