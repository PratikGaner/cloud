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
	"log"
	"os"

	"github.com/imdario/mergo"
	"github.com/pkg/errors"

	kantocfg "github.com/eclipse-kanto/suite-connector/config"
	"github.com/eclipse-kanto/suite-connector/logger"

	"github.com/eclipse-kanto/azure-connector/cmd/azure-connector/app"
	azureflags "github.com/eclipse-kanto/azure-connector/flags"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers/passthrough"

	mapperconfig "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/config"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/handlers/command"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/handlers/telemetry"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/protobuf"

	azurecfg "github.com/eclipse-kanto/azure-connector/config"
)

var (
	version = "development"
)

func main() {
	f := flag.NewFlagSet("azure-connector", flag.ContinueOnError)

	cmd := &AzureSettingsExt{
		AzureSettings: &azurecfg.AzureSettings{},
	}
	azureflags.Add(f, cmd.AzureSettings)
	addMessageHandlers(f, cmd)
	fConfigFile := azureflags.AddGlobal(f)

	if err := azureflags.Parse(f, os.Args[1:], version, os.Exit); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		} else {
			os.Exit(2)
		}
	}

	settings := defaultSettings()
	if err := kantocfg.ReadConfig(*fConfigFile, settings); err != nil {
		log.Fatal(errors.Wrap(err, "cannot parse config"))
	}

	cli := azureflags.Copy(f)
	if err := mergo.Map(settings, cli, mergo.WithOverwriteWithEmptyValue); err != nil {
		log.Fatal(errors.Wrap(err, "cannot process settings"))
	}

	if err := settings.Validate(); err != nil {
		log.Fatal(errors.Wrap(err, "settings validation error"))
	}

	loggerOut, logger := logger.Setup("azure-connector", &settings.LogSettings)
	defer loggerOut.Close()

	logger.Infof("Starting azure connector %s", version)
	azureflags.ConfigCheck(logger, *fConfigFile)

	mapperConfig, err := mapperconfig.LoadMessageMapperConfig(settings.MessageMapperConfig)
	if err != nil {
		logger.Error("cannot load message mapper config", err, nil)
	}
	marshaller := protobuf.NewProtobufJSONMarshaller(mapperConfig)
	telemetryHandlers := createTelemetryHandlers(settings, mapperConfig, marshaller)
	commandHandlers := createCommandHandlers(settings, mapperConfig, marshaller)

	if err := app.MainLoop(settings.AzureSettings, logger, nil, telemetryHandlers, commandHandlers); err != nil {
		logger.Error("Init failure", err, nil)

		loggerOut.Close()

		os.Exit(1)
	}
}

func createTelemetryHandlers(settings *AzureSettingsExt, mapperConfig *mapperconfig.MessageMapperConfig, marshaller protobuf.Marshaller) []handlers.TelemetryHandler {
	handlers := []handlers.TelemetryHandler{}
	passthroughHandler := passthrough.CreateTelemetryHandler(settings.PassthroughDeviceTopics)
	handlers = append(handlers, passthroughHandler)
	if mapperConfig != nil {
		thingsHandler := telemetry.CreateThingsTelemetryHandler(mapperConfig, marshaller)
		handlers = append(handlers, thingsHandler)
	}
	return handlers
}

func createCommandHandlers(settings *AzureSettingsExt, mapperConfig *mapperconfig.MessageMapperConfig, marshaller protobuf.Marshaller) []handlers.CommandHandler {
	handlers := []handlers.CommandHandler{}
	passthroughHandler := command.CreatePassthroughCommandHandler(settings.PassthroughCommandNames)
	handlers = append(handlers, passthroughHandler)
	if mapperConfig != nil {
		thingsHandler := command.CreateThingsCommandHandler(mapperConfig, marshaller)
		handlers = append(handlers, thingsHandler)
	}
	return handlers
}
