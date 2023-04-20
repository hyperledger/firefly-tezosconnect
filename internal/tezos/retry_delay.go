package tezos

import (
	"context"
	"time"

	"github.com/hyperledger/firefly-common/pkg/log"
)

func (c *tezosConnector) doFailureDelay(ctx context.Context, failureCount int) bool {
	if failureCount <= 0 {
		return false
	}

	retryDelay := c.retry.InitialDelay
	for i := 0; i < (failureCount - 1); i++ {
		retryDelay = time.Duration(float64(retryDelay) * c.retry.Factor)
		if retryDelay > c.retry.MaximumDelay {
			retryDelay = c.retry.MaximumDelay
			break
		}
	}
	log.L(ctx).Debugf("Retrying after %.2f (failures=%d)", retryDelay.Seconds(), failureCount)
	select {
	case <-time.After(retryDelay):
		return false
	case <-ctx.Done():
		return true
	}
}
