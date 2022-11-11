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

package command

import (
	"fmt"
	"strings"

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"
	"github.com/eclipse-kanto/azure-connector/util"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pkg/errors"
)

const commandPassthroughHandlerName = "command_passthrough_handler"

type passthroughCommandHandler struct {
	commandNames []string
}

// CreatePassthroughCommandHandler instantiates a passthrough command message handler
func CreatePassthroughCommandHandler(commandNames string) handlers.CommandHandler {
	return &passthroughCommandHandler{
		commandNames: strings.Split(commandNames, ","),
	}
}

func (h *passthroughCommandHandler) Init(connInfo *config.RemoteConnectionInfo) error {
	return nil
}

func (h *passthroughCommandHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	cloudMessage, err := parseCommandMessage(msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot deserialize cloud message")
	}
	if util.ContainsString(h.commandNames, cloudMessage.CommandName) {
		msg.SetContext(connector.SetTopicToCtx(msg.Context(), cloudMessage.ApplicationID+"/"+cloudMessage.CommandName))
		return []*message.Message{msg}, nil
	}
	return nil, fmt.Errorf("cloud command name '%s' is not supported", cloudMessage.CommandName)
}

func (h *passthroughCommandHandler) Name() string {
	return commandPassthroughHandlerName
}
