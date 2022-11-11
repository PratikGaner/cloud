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

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	mapperconfig "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/config"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/protobuf"

	"github.com/eclipse/ditto-clients-golang/protocol"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThingsMessageHandler(t *testing.T) {
	messageHandler := &thingsCommandHandler{}

	messageHandler.Init(&config.RemoteConnectionInfo{})
	assert.Equal(t, commandThingsHandlerName, messageHandler.Name())
}

func TestHandleC2DMessageCorrectly(t *testing.T) {
	handler := createThingsCommandHandler(t)
	var testData = []struct {
		jsonPayload   string
		msgTopic      string
		topic         string
		path          string
		correlationID string
	}{
		{
			`{
				"appId": "app1",
				"cmdName": "container.manifest",
				"cId": "C2D-msg-correlation-id",
				"eVer": "2.0",
				"pVer": "1.0",
				"p": "Cq4BCglzb21lLW5hbWUSBTAuMC4xGmAKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaHQoLdGVzdC1kb21haW4SBlZBUjE9MhIGVkFSPXR6KgkKB1JVTk5JTkcaOAoQc2Vjb25kLWNvbnRhaW5lchIkZG9ja2VyLmlvL2xpYnJhcnkvaGVsbG8td29ybGQ6bGF0ZXN0"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req/C2D-msg-correlation-id/apply",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/apply",
			"/features/ContainerOrchestrator/inbox/messages/apply",
			"C2D-msg-correlation-id",
		},
		{
			`{
				"appId": "app1",
				"cmdName": "container.manifest",
				"cId": "C2D-msg-correlation-id2",
				"eVer": "2.0",
				"pVer": "1.0",
				"p": "Cq4BCglzb21lLW5hbWUSBTAuMC4xGmAKEW15LWNvbnRhaW5lci1uYW1lEiFkb2NrZXIuaW8vbGlicmFyeS9pbmZsdXhkYjpsYXRlc3QaHQoLdGVzdC1kb21haW4SBlZBUjE9MhIGVkFSPXR6KgkKB1JVTk5JTkcaOAoQc2Vjb25kLWNvbnRhaW5lchIkZG9ja2VyLmlvL2xpYnJhcnkvaGVsbG8td29ybGQ6bGF0ZXN0"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req/C2D-msg-correlation-id2/apply",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/apply",
			"/features/ContainerOrchestrator/inbox/messages/apply",
			"C2D-msg-correlation-id2",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.msgTopic, func(t *testing.T) {
			azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			azureMsg := azureMessages[0]
			azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

			c2dMessage := &protocol.Envelope{}
			err = json.Unmarshal(azureMsg.Payload, c2dMessage)
			require.NoError(t, err)

			assert.Equal(t, testValues.msgTopic, azureMsgTopic)
			assert.Equal(t, testValues.topic, c2dMessage.Topic.String())
			assert.Equal(t, testValues.path, c2dMessage.Path)
			assert.Equal(t, "application/json", c2dMessage.Headers.ContentType())
			assert.Equal(t, testValues.correlationID, c2dMessage.Headers.CorrelationID())
			assert.True(t, len(c2dMessage.Value.(map[string]interface{})) > 0)
		})
	}
}

func TestValidC2DMessageWithTypeSimpleMessage(t *testing.T) {
	handler := createThingsCommandHandler(t)
	var testSimpleData = []struct {
		jsonPayload    string
		msgTopic       string
		topic          string
		path           string
		correlationID  string
		payloadText    string
		payloadVersion string
	}{
		{
			`{
				"appId": "app1",
				"cmdName": "simple.message",
				"cId": "C2D-msg-correlation-id",
				"eVer": "2.0",
				"pVer": "1.0",
				"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req/C2D-msg-correlation-id/send",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/send",
			"/features/ContainerOrchestrator/inbox/messages/send",
			"C2D-msg-correlation-id",
			"simple text added",
			"1.0",
		},
		{
			`{
				"appId": "app1",
				"cmdName": "simple.message",
				"eVer": "2.0",
				"pVer": "1.0",
				"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
			}`,
			"command//azure.edge:dummy-hub:dummy-device:edge:containers/req//send",
			"azure.edge/dummy-hub:dummy-device:edge:containers/things/live/messages/send",
			"/features/ContainerOrchestrator/inbox/messages/send",
			"",
			"simple text added",
			"1.0",
		},
	}

	for _, testValues := range testSimpleData {
		t.Run(testValues.msgTopic, func(t *testing.T) {
			azureMessages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			azureMsg := azureMessages[0]
			azureMsgTopic, _ := connector.TopicFromCtx(azureMsg.Context())

			c2dMessage := &protocol.Envelope{}
			err = json.Unmarshal(azureMsg.Payload, c2dMessage)
			require.NoError(t, err)

			assert.Equal(t, testValues.msgTopic, azureMsgTopic)
			assert.Equal(t, testValues.topic, c2dMessage.Topic.String())
			assert.Equal(t, testValues.path, c2dMessage.Path)
			assert.Equal(t, "application/json", c2dMessage.Headers.ContentType())
			assert.Equal(t, testValues.correlationID, c2dMessage.Headers.CorrelationID())
			assert.True(t, len(c2dMessage.Value.(map[string]interface{})) > 0)
		})
	}
}

func TestValidC2DMessageWithTypeUnsupportedProtobufFieldInJSON(t *testing.T) {
	handler := createThingsCommandHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "unsupported.field",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"pVer": "1.0",
		"p": "ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44"
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestUnsupportedC2DVSSMessageType(t *testing.T) {
	handler := createThingsCommandHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "subscribeCommand",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"pVer": "1.0",
		"p": {}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestUnsupportedC2DMessageType(t *testing.T) {
	handler := createThingsCommandHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.non-existing",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"pVer": "1.0",
		"p": {}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestInvalidPayloadC2DMessageType(t *testing.T) {
	handler := createThingsCommandHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"pVer": "1.0",
		"p": "Cq4BCglzb21lLW5hbWUSBTAuMdGVzdC1kb21haW4SBlZBUjE9MhIGVkFSPXR6KgkKB1JVTk5JTkcaOAoQc2Vjb25kLWNvbnRhaW5lchIkZG9ja2VyLmlvL2xpYnJhcnkvaGVsbG8td29ybGQ6bGF0ZXN0"
	}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNoPayloadC2DMessageType(t *testing.T) {
	handler := createThingsCommandHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": ""
	}`
	messages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.NoError(t, err)
	dittoMessage := &protocol.Envelope{}
	err = json.Unmarshal(messages[0].Payload, dittoMessage)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{}, dittoMessage.Value.(map[string]interface{}))
}

func TestNoProtoFileC2DMessageType(t *testing.T) {
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "message.no.proto.file",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": "dummy_payload"
	}`
	expectedPayload := "dummy_payload"
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)
}

func TestNoProtoFileManifest(t *testing.T) {
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": "dummy_payload"
	}`
	expectedPayload := `{"manifest":"dummy_payload"}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": false
	}`
	expectedPayload = `{"manifest":false}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": 1
	}`
	expectedPayload = `{"manifest":1}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": 10.6
	}`
	expectedPayload = `{"manifest":10.6}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": {"dummy_array":[{"dummy_id": "id_1", "dummy_name": "name_1"},{"dummy_id": "id_2", "dummy_name": "name_2"}]}
	}`
	expectedPayload = `{"manifest":{"dummy_array":[{"dummy_id":"id_1","dummy_name":"name_1"},{"dummy_id":"id_2","dummy_name":"name_2"}]}}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)
}

func TestNoProtoFileRetainCorrelationIdC2DMessageType(t *testing.T) {
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.retain.correlation.id",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": "dummy_payload"
	}`
	expectedPayload := `{"correlationId":"C2D-msg-correlation-id","payload":"dummy_payload"}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.retain.correlation.id",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": 1
	}`
	expectedPayload = `{"correlationId":"C2D-msg-correlation-id","payload":1}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.retain.correlation.id",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": 71.72
	}`
	expectedPayload = `{"correlationId":"C2D-msg-correlation-id","payload":71.72}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.retain.correlation.id",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": false
	}`
	expectedPayload = `{"correlationId":"C2D-msg-correlation-id","payload":false}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)

	jsonPayload = `{
		"appId": "app1",
		"cmdName": "message.no.proto.file.retain.correlation.id",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"p": {"x":"y"}
	}`
	expectedPayload = `{"correlationId":"C2D-msg-correlation-id","payload":{"x":"y"}}`
	assertNoProtoFilePayload(t, jsonPayload, expectedPayload)
}

func assertNoProtoFilePayload(t *testing.T, jsonPayload, expectedPayload string) {
	handler := createThingsCommandHandler(t)
	messages, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.NoError(t, err)
	dittoMessage := &protocol.Envelope{}
	err = json.Unmarshal(messages[0].Payload, dittoMessage)
	require.NoError(t, err)
	switch dittoValue := dittoMessage.Value.(type) {
	case string:
		assert.Equal(t, expectedPayload, dittoValue)
	case map[string]interface{}:
		byteValue, err := json.Marshal(dittoMessage.Value.(map[string]interface{}))
		require.NoError(t, err)
		assert.Equal(t, expectedPayload, string(byteValue))
	}
}

func TestInvalidJSONC2DMessageType(t *testing.T) {
	handler := createThingsCommandHandler(t)
	jsonPayload := `{
		"appId": "app1",
		"cmdName": "container.manifest",
		"cId": "C2D-msg-correlation-id",
		"eVer": "2.0",
		"pVer": "1.0",
		"p": {}`
	_, err := handler.HandleMessage(createWatermillMessageForC2D([]byte(jsonPayload)))
	require.Error(t, err)
}

func createThingsCommandHandler(t *testing.T) handlers.CommandHandler {
	mapperConfig, _ := mapperconfig.LoadMessageMapperConfig("../internal/testdata/handlers-mapper-config.json")
	messageHandler := CreateThingsCommandHandler(mapperConfig, protobuf.NewProtobufJSONMarshaller(mapperConfig))
	messageHandler.Init(&config.RemoteConnectionInfo{DeviceID: "dummy-device", HubName: "dummy-hub"})
	return messageHandler
}
