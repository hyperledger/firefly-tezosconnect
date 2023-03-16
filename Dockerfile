FROM golang:1.19-buster AS builder
ARG BUILD_VERSION
ENV BUILD_VERSION=${BUILD_VERSION}
ADD . /tezosconnect
WORKDIR /tezosconnect
RUN make

FROM debian:buster-slim
WORKDIR /tezosconnect
RUN apt update -y \
 && apt install -y curl jq \
 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /tezosconnect/firefly-tezosconnect /usr/bin/tezosconnect

ENTRYPOINT [ "/usr/bin/tezosconnect" ]
