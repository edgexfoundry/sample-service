#
# Copyright (c) 2018, 2019 Intel
#
# SPDX-License-Identifier: Apache-2.0
#
FROM golang:1.11.8-alpine

ENV GO111MODULE=on
ENV GOCACHE=/tmp/.cache

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk update && apk add make git

WORKDIR /go/src/github.com/edgexfoundry

COPY go.mod .

RUN go mod download