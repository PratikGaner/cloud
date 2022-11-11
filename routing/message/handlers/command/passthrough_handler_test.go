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
	"encoding/json"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	"github.com/eclipse-kanto/suite-connector/connector"

	routingmessage "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TesCreatetPassthroughMessageHandler(t *testing.T) {
	messageHandler := &passthroughCommandHandler{}

	messageHandler.Init(&config.RemoteConnectionInfo{})
	assert.Equal(t, commandPassthroughHandlerName, messageHandler.Name())
}

func TestValidC2DMessageWithSupportedCommandName(t *testing.T) {
	handler := createCommandHandler(t)
	jsonPayload := `{
		"appId": "datapoints",
		"cId": "C2D-msg-correlation-id",
		"cmdName": "testCommand",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": "CoEBCglzb21lLW5hbWUSBTAuMC4xGjYKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaNQoQc2Vjb25kLWNvbnRhaW5lchIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0"
	}`

	azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.NoError(t, err)

	azureMsg := azureMessages[0]
	azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

	c2dMessage := &routingmessage.CloudMessage{}
	err = json.Unmarshal(azureMsg.Payload, c2dMessage)
	require.NoError(t, err)

	assert.Equal(t, "datapoints/testCommand", azureMsgTopic)
	assert.Equal(t, "testCommand", c2dMessage.CommandName)
	assert.Equal(t, "datapoints", c2dMessage.ApplicationID)
	assert.Equal(t, "C2D-msg-correlation-id", c2dMessage.CorrelationID)
	assert.Equal(t, "1.0.0", c2dMessage.EnvelopeVersion)
	assert.True(t, len(c2dMessage.Payload.(string)) > 0)
}

func TestValidC2DMessageWithNotSupportedCommandName(t *testing.T) {
	handler := createCommandHandler(t)
	jsonPayload := `{
		"appId": "datapoints",
		"cId": "C2D-msg-correlation-id",
		"cmdName": "not-supported",
		"eVer": "1.0.0",
		"pVer": "1.0.0",
		"p": "CoEBCglzb21lLW5hbWUSBTAuMC4xGjYKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaNQoQc2Vjb25kLWNvbnRhaW5lchIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0"
	}`

	azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
	assert.Nil(t, azureMessages)
}

func TestInvalidC2DMessagePayload(t *testing.T) {
	handler := createCommandHandler(t)
	jsonPayload := "invalid-payload"
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func createCommandHandler(t *testing.T) handlers.CommandHandler {
	messageHandler := &passthroughCommandHandler{
		commandNames: []string{"testVal", "testCommand"},
	}
	messageHandler.Init(&config.RemoteConnectionInfo{DeviceID: "dummy-device", HubName: "dummy-hub"})
	return messageHandler
}

func createWatermillMessageForC2D(payload []byte) *message.Message {
	message := message.NewMessage(watermill.NewUUID(), payload)
	cloudMessage := &routingmessage.CloudMessage{}
	json.Unmarshal(payload, cloudMessage)
	return message
}
