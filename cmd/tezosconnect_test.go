package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-tezosconnect/internal/tezos"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	testCases := []struct {
		name        string
		errMsg      string
		initFunc    func()
		cleanupFunc func()
	}{
		{
			name: "success",
			initFunc: func() {
				f, err := os.Create("firefly.tezosconnect")
				assert.NoError(t, err)
				err = f.Close()
				assert.NoError(t, err)

				dir, err := os.MkdirTemp("", "ldb_*")
				assert.NoError(t, err)
				config.Set("persistence.leveldb.path", dir)
			},
			cleanupFunc: func() {
				err := os.Remove("firefly.tezosconnect")
				assert.NoError(t, err)
			},
		},
		{
			name:        "error on config not found",
			initFunc:    func() {},
			cleanupFunc: func() {},
			errMsg:      "FF00101: Failed to read config: Config File \"firefly.tezosconnect\" Not Found",
		},
		{
			name: "error on NewTezosConnector",
			initFunc: func() {
				f, err := os.Create("firefly.tezosconnect")
				assert.NoError(t, err)
				err = f.Close()
				assert.NoError(t, err)

				connectorConfig.Set(tezos.TxCacheSize, "-1")
			},
			cleanupFunc: func() {
				err := os.Remove("firefly.tezosconnect")
				assert.NoError(t, err)
			},
			errMsg: "FF23040: Failed to initialize transaction cache: Must provide a positive size",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.initFunc()
			ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second)

			err := run(ctx, cancelCtx)

			tc.cleanupFunc()

			if tc.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
