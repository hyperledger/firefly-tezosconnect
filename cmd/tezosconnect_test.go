package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	InitConfig()
	testCases := []struct {
		name    string
		errMsg  string
		cfgFile string
	}{
		{
			name:    "success",
			cfgFile: "../test/firefly.tezosconnect.yaml",
		},
		{
			name:    "error on config not found",
			cfgFile: "../test/missing.firefly.tezosconnect.yaml",
			errMsg:  "FF00101",
		},
		{
			name:    "wrong persistence type error",
			cfgFile: "../test/bad-persistence-type.tezosconnect.yaml",
			errMsg:  "FF21043",
		},
		{
			name:    "error on NewTezosConnector",
			cfgFile: "../test/no-connector.tezosconnect.yaml",
			errMsg:  "FF23051",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfgFile = tc.cfgFile
			ctx, cancelCtx := context.WithTimeout(context.Background(), time.Second)

			err := run(ctx, cancelCtx)

			if tc.errMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunMigrationsBadConfig(t *testing.T) {
	rootCmd.SetArgs([]string{
		"migrate", "leveldb2postgres", "-f", "../test/no-connector.tezosconnect.yaml",
	})
	defer rootCmd.SetArgs([]string{})

	err := Execute()
	assert.Regexp(t, "FF21049", err)
}