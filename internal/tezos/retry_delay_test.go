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

	c.retry.MaximumDelay = 1 * time.Microsecond
	c.retry.InitialDelay = 100 * time.Microsecond

	testCases := []struct {
		name         string
		failureCount int
		result       bool
	}{
		{
			name:         "zero failure count",
			failureCount: 0,
			result:       false,
		},
		{
			name:         "retry delay exceeds max delay",
			failureCount: 10,
			result:       false,
		},
		{
			name:         "ctx done",
			failureCount: 0,
			result:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := c.doFailureDelay(context.Background(), 0)
			assert.Equal(t, tc.result, res)
		})
	}
}
