VGO=go
GOFILES := $(shell find cmd internal -name '*.go' -print)

firefly-tezosconnect: ${GOFILES}
		$(VGO) build -o ./firefly-tezosconnect -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=$(BUILD_VERSION)" -tags=prod -v ./tezosconnect 