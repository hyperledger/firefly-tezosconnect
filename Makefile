VGO=go
GOFILES := $(shell find cmd internal -name '*.go' -print)
GOBIN := $(shell $(VGO) env GOPATH)/bin
LINT := $(GOBIN)/golangci-lint

.DELETE_ON_ERROR:

all: build test go-mod-tidy
test: deps
		$(VGO) test ./internal/... ./cmd/... -cover -coverprofile=coverage.txt -covermode=atomic -timeout=30s
coverage.html:
		$(VGO) tool cover -html=coverage.txt
coverage: test coverage.html
lint: ${LINT}
		$(LINT) run -v --timeout 5m
${LINT}:
		$(VGO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.47.3


firefly-tezosconnect: ${GOFILES}
		$(VGO) build -o ./firefly-tezosconnect -ldflags "-X main.buildDate=`date -u +\"%Y-%m-%dT%H:%M:%SZ\"` -X main.buildVersion=$(BUILD_VERSION)" -tags=prod -v ./tezosconnect
go-mod-tidy: .ALWAYS
		$(VGO) mod tidy
build: firefly-tezosconnect
.ALWAYS: ;
clean:
		$(VGO) clean
deps:
		$(VGO) get ./tezosconnect
docker:
		docker build --build-arg BUILD_VERSION=${BUILD_VERSION} ${DOCKER_ARGS} -t oneof/firefly-tezosconnect .