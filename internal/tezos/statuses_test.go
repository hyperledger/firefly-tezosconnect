package tezos

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLive(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	status, _, err := c.IsLive(ctx)
	assert.NoError(t, err)
	assert.True(t, status.Up)
}

func TestIsReady(t *testing.T) {
	ctx, c, _, done := newTestConnector(t)
	defer done()

	status, _, err := c.IsReady(ctx)
	assert.NoError(t, err)
	assert.True(t, status.Ready)
}
