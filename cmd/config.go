// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Apache License 2.0 which is available at
// https://www.apache.org/licenses/LICENSE-2.0
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"

	"github.com/eclipse-kanto/azure-connector/config"
)

const (
	defaultMessageMapperConfig = "message-mapper-config.json"

	flagMessageMapperConfig     = "messageMapperConfig"
	flagPassthroughDeviceTopics = "passthroughDeviceTopics"
	flagPassthroughCommandNames = "passthroughCommandNames"
)

// AzureSettingsExt wraps the general configurable data of the Cloud Connector with with custom properties
type AzureSettingsExt struct {
	PassthroughDeviceTopics string
	PassthroughCommandNames string
	MessageMapperConfig     string
	*config.AzureSettings
}

func defaultSettings() *AzureSettingsExt {
	return &AzureSettingsExt{
		MessageMapperConfig: defaultMessageMapperConfig,
		AzureSettings:       config.DefaultSettings(),
	}
}

func addMessageHandlers(f *flag.FlagSet, settings *AzureSettingsExt) {
	def := defaultSettings()

	f.StringVar(&settings.MessageMapperConfig,
		flagMessageMapperConfig, def.MessageMapperConfig,
		"The path to the configuration file for the message mappings",
	)

	f.StringVar(&settings.PassthroughDeviceTopics,
		flagPassthroughDeviceTopics, def.PassthroughDeviceTopics,
		"List of passthrough device topics that the cloud connector subscribes for and forwards messages to the Azure IoT Hub",
	)

	f.StringVar(&settings.PassthroughCommandNames,
		flagPassthroughCommandNames, def.PassthroughCommandNames,
		"List of passthrough command names that the cloud connector filters and forwards inside the device",
	)
}
