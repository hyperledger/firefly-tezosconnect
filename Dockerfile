FROM golang:1.21-alpine3.19 AS builder
ARG BUILD_VERSION
ENV BUILD_VERSION=${BUILD_VERSION}
ADD . /tezosconnect
WORKDIR /tezosconnect
RUN make

# Copy the migrations from FFTM down into our local migrations directory
RUN DB_MIGRATIONS_DIR=$(go list -f '{{.Dir}}' github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi | sed 's|pkg/ffcapi|db|') \
 && cp -R $DB_MIGRATIONS_DIR db

FROM debian:buster-slim
WORKDIR /tezosconnect
RUN apt update -y \
 && apt install -y curl jq \
 && rm -rf /var/lib/apt/lists/* \
 && curl -sL "https://github.com/golang-migrate/migrate/releases/download/$(curl -sL https://api.github.com/repos/golang-migrate/migrate/releases/latest | jq -r '.name')/migrate.linux-amd64.tar.gz" | tar xz \
 && chmod +x ./migrate \
 && mv ./migrate /usr/bin/migrate
COPY --from=builder /tezosconnect/firefly-tezosconnect /usr/bin/tezosconnect
COPY --from=builder /tezosconnect/db/ /tezosconnect/db/

ENTRYPOINT [ "/usr/bin/tezosconnect" ]
