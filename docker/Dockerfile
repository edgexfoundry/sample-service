#
# Copyright (c) 2018, 2019 Intel
#
# SPDX-License-Identifier: Apache-2.0
#
FROM golang:1.11-alpine

ENV GO111MODULE=on

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk update && apk add make git curl gcc libc-dev

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(go env GOPATH)/bin v1.15.0

WORKDIR /go/src/github.com/edgexfoundry

COPY go.mod .

RUN go mod download