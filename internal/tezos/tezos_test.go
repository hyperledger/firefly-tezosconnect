package tezos

import (
	"context"
	"testing"

	"github.com/OneOf-Inc/firefly-tezosconnect/mocks/rpcbackendmocks"
	"github.com/hyperledger/firefly-common/pkg/config"
	"github.com/hyperledger/firefly-common/pkg/ffresty"
	"github.com/hyperledger/firefly-common/pkg/fftls"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func newTestConnector(t *testing.T) (context.Context, *tezosConnector, *rpcbackendmocks.Backend, func()) {
	mRPC := &rpcbackendmocks.Backend{}
	config.RootConfigReset()
	conf := config.RootSection("unittest")
	InitConfig(conf)
	conf.Set(ffresty.HTTPConfigURL, "http://localhost:8545")
	conf.Set(BlockPollingInterval, "1h") // Disable for tests that are not using it
	logrus.SetLevel(logrus.DebugLevel)
	ctx, done := context.WithCancel(context.Background())
	cc, err := NewTezosConnector(ctx, conf)
	assert.NoError(t, err)
	c := cc.(*tezosConnector)
	c.backend = mRPC

	return ctx, c, mRPC, func() {
		done()
		mRPC.AssertExpectations(t)
		c.WaitClosed()
	}
}

func TestConnectorInit(t *testing.T) {
	config.RootConfigReset()
	conf := config.RootSection("unittest")
	InitConfig(conf)

	cc, err := NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23025", err)

	conf.Set(ffresty.HTTPConfigURL, "http://localhost:8545")
	conf.Set(EventsCatchupThreshold, 1)
	conf.Set(EventsCatchupPageSize, 500)

	cc, err = NewTezosConnector(context.Background(), conf)
	assert.NoError(t, err)
	assert.Equal(t, int64(500), cc.(*tezosConnector).catchupThreshold) // set to page size

	tlsConf := conf.SubSection("tls")
	tlsConf.Set(fftls.HTTPConfTLSEnabled, true)
	tlsConf.Set(fftls.HTTPConfTLSCAFile, "!!!badness")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF00153", err)
	tlsConf.Set(fftls.HTTPConfTLSEnabled, false)

	conf.Set(ConfigDataFormat, "wrong")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23032.*wrong", err)

	conf.Set(ConfigDataFormat, "map")
	conf.Set(BlockCacheSize, "-1")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23040", err)

	conf.Set(BlockCacheSize, "1")
	conf.Set(TxCacheSize, "-1")
	cc, err = NewTezosConnector(context.Background(), conf)
	assert.Regexp(t, "FF23040", err)
}
