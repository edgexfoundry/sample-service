// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (c) 2020 Intel
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/edgexfoundry/device-sdk-go/pkg/startup"
	"github.com/edgexfoundry/sample-service"
	"github.com/edgexfoundry/sample-service/driver"
)

const (
	serviceName string = "sample-service"
)

func main() {
	d := driver.NewProtocolDriver()
	startup.Bootstrap(serviceName, sample_service.Version, d)
}
