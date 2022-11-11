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

package telemetry

import (
	"encoding/json"
	"testing"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	routingmessage "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message"
	mapperconfig "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/config"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/protobuf"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	commonMessageMapperConfig            = "../internal/testdata/handlers-mapper-config.json"
	convertDittoValueMessageMapperConfig = "../internal/testdata/convert-ditto-value-mappings.json"
)

func TestThingsMessageHandler(t *testing.T) {
	messageHandler := &thingsTelemetryHandler{}

	messageHandler.Init(&config.RemoteConnectionInfo{})
	assert.Equal(t, telemetryHandlerName, messageHandler.Name())
	assert.Equal(t, "event/#,e/#,telemetry/#,t/#", messageHandler.Topics())
}

func TestHandleContainerMessageTypes(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	var testData = []struct {
		jsonPayload     string
		messageType     int
		messageSubType  string
		correlationID   string
		protobufPayload string
	}{
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/created",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/created",
				"value": {
				  "name": "influxdb",
				  "imageRef": "docker.io/library/influxdb:latest",
				  "config": {
					"domainName": "some-domain",
					"restartPolicy": {
					  "type": "UNLESS_STOPPED"
					},
					"log": {
					  "type": "JSON_FILE",
					  "maxFiles": 2,
					  "maxSize": "100M",
					  "mode": "BLOCKING"
					}
				  },
				  "createdAt": "2021-06-03T11:52:56.614763386Z"
				}
			}`,
			1,
			"container.created",
			"",
			"CghpbmZsdXhkYhIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0Gj4KC3NvbWUtZG9tYWluMhAaDlVOTEVTU19TVE9QUEVEWh0KCUpTT05fRklMRRACGgQxMDBNKghCTE9DS0lORyIeMjAyMS0wNi0wM1QxMTo1Mjo1Ni42MTQ3NjMzODZa",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/created",
				"headers": {
				  "response-required": false,
				  "correlation-id": "some-correlation-id"
				},
				"path": "/features/ContainerOrchestator/outbox/messages/created",
				"value": {
				  "name": "influxdb",
				  "imageRef": "docker.io/library/influxdb:latest",
				  "config": {
					"domainName": "some-domain",
					"restartPolicy": {
					  "type": "UNLESS_STOPPED"
					},
					"log": {
					  "type": "JSON_FILE",
					  "maxFiles": 2,
					  "maxSize": "100M",
					  "mode": "BLOCKING"
					}
				  },
				  "createdAt": "2021-06-03T11:52:56.614763386Z"
				}
			}`,
			1,
			"container.created",
			"some-correlation-id",
			"CghpbmZsdXhkYhIhZG9ja2VyLmlvL2xpYnJhcnkvaW5mbHV4ZGI6bGF0ZXN0Gj4KC3NvbWUtZG9tYWluMhAaDlVOTEVTU19TVE9QUEVEWh0KCUpTT05fRklMRRACGgQxMDBNKghCTE9DS0lORyIeMjAyMS0wNi0wM1QxMTo1Mjo1Ni42MTQ3NjMzODZa",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/removed",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/removed",
				"value": {
				  "name": "influxdb"
				}
			}`,
			1,
			"container.removed",
			"",
			"CghpbmZsdXhkYg==",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/stateChanged",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/stateChanged",
				"value": {
				  "name": "influxdb",
				  "state": {
					"status": "RUNNING",
					"pid": 4294967295,
					"startedAt": "2021-04-29T17:37:47.15018946Z"
				  }
				}
			}`,
			1,
			"container.stateChanged",
			"",
			"CghpbmZsdXhkYhIuCgdSVU5OSU5HEP////8PKh0yMDIxLTA0LTI5VDE3OjM3OjQ3LjE1MDE4OTQ2Wg==",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/twin/commands/modify",
				"headers": {
				  "response-required": false
				},
				"path": "/features/ContainerOrchestrator/properties/state/status",
				"value": {
				  "manifest": {
					"name": "some-name",
					"version": "0.0.1"
				  },
				  "state": "FINISHED_ERROR",
				  "error": {
					"code": 500,
					"message": "something went wrong :("
				  }
				}
			}`,
			1,
			"container.manifest",
			"",
			"ChIKCXNvbWUtbmFtZRIFMC4wLjESDkZJTklTSEVEX0VSUk9SGhwI9AMSF3NvbWV0aGluZyB3ZW50IHdyb25nIDoo",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/simplySend",
				"headers": {
					"response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/simplySend",
				"value": {
					"message_id": "some-message-id-1234",
					"text": "simple text added",
					"version": "1.0.8"
				}
			}`,
			1,
			"simple.message",
			"",
			"ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			d2cMessage := &routingmessage.TelemetryMessage{}
			err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
			require.NoError(t, err)

			protobufPayload, ok := d2cMessage.Payload.(string)
			assert.True(t, ok)

			assert.Equal(t, "", d2cMessage.ApplicationID)
			assert.Equal(t, testValues.messageType, d2cMessage.MessageType)
			assert.Equal(t, testValues.messageSubType, d2cMessage.MessageSubType)
			assert.Equal(t, testValues.correlationID, d2cMessage.CorrelationID)
			assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
			assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
			assert.Equal(t, testValues.protobufPayload, protobufPayload)
		})
	}
}

func TestOptimizeDittoPayload(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	var testData = []struct {
		jsonPayload     string
		messageType     int
		messageSubType  string
		correlationID   string
		protobufPayload string
	}{
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/singleFieldArray",
				"headers": {
				  "response-required": false
				},
				"path": "/outbox/messages/singleFieldArray",
				"value": [
					{
						"message_id": "messageId1",
						"text": "dummy_text",
						"version": "1.0.0"
					},
					{
						"message_id": "messageId2",
						"text": "dummy_text",
						"version": "1.0.0"
					},
					{
						"message_id": "messageId3",
						"text": "dummy_text",
						"version": "1.0.0"
					}
				]
			}`,
			1,
			"single.field.array",
			"",
			"Ch8KCm1lc3NhZ2VJZDESCmR1bW15X3RleHQaBTEuMC4wCh8KCm1lc3NhZ2VJZDISCmR1bW15X3RleHQaBTEuMC4wCh8KCm1lc3NhZ2VJZDMSCmR1bW15X3RleHQaBTEuMC4w",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/singleFieldIntArray",
				"headers": {
				  "response-required": false
				},
				"path": "/outbox/messages/singleFieldIntArray",
				"value": 1
			}`,
			1,
			"single.field.int.array",
			"",
			"CgEB",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/singleFieldBoolArray",
				"headers": {
				  "response-required": false
				},
				"path": "/outbox/messages/singleFieldBoolArray",
				"value": false
			}`,
			1,
			"single.field.bool.array",
			"",
			"CgEA",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/singleFieldStringArray",
				"headers": {
				  "response-required": false
				},
				"path": "/outbox/messages/singleFieldStringArray",
				"value": "dummy_value"
			}`,
			1,
			"single.field.string.array",
			"",
			"CgtkdW1teV92YWx1ZQ==",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			d2cMessage := &routingmessage.TelemetryMessage{}
			err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
			require.NoError(t, err)

			protobufPayload, ok := d2cMessage.Payload.(string)
			assert.True(t, ok)

			assert.Equal(t, "", d2cMessage.ApplicationID)
			assert.Equal(t, testValues.messageType, d2cMessage.MessageType)
			assert.Equal(t, testValues.messageSubType, d2cMessage.MessageSubType)
			assert.Equal(t, testValues.correlationID, d2cMessage.CorrelationID)
			assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
			assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
			assert.Equal(t, testValues.protobufPayload, protobufPayload)
		})
	}
}

func TestConvertDittoValue(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	var testData = []struct {
		testName       string
		jsonPayload    string
		messageType    int
		messageSubType string
		convertedValue string
	}{
		{
			"initial.increment.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/increment.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/increment.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"message": "dummy_message"
				}
			}`,
			1,
			"increment.mapping",
			"{\"counter\":1}",
		},
		{
			"increment.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/increment.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/increment.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"message": "dummy_message"
				}
			}`,
			1,
			"increment.mapping",
			"{\"counter\":2}",
		},
		{
			"static.value.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/static.value.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/static.value.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"message": "dummy_message"
				}
			}`,
			1,
			"static.value.mapping",
			"{\"bool.key\":true,\"int.key\":7,\"string.key\":\"dummy_value\"}",
		},
		{
			"reference.value.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/simple.ref.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/simple.ref.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"int_key" : 7,
					"bool_key": true,
					"string_key": "ref_value"
				}
			}`,
			1,
			"simple.ref.mapping",
			"{\"bool.key\":true,\"int.key\":7,\"string.key\":\"ref_value\"}",
		},
		{
			"skip.value.missing.reference.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/simple.ref.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/simple.ref.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"bool_key": true,
					"string_key": "ref_value"
				}
			}`,
			1,
			"simple.ref.mapping",
			"{\"bool.key\":true,\"string.key\":\"ref_value\"}",
		},
		{
			"static.nested.value.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/static.nested.value.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/static.nested.value.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"message": "dummy_message"
				}
			}`,
			1,
			"static.nested.value.mapping",
			"{\"nested.object\":{\"bool.key\":true,\"int.key\":7,\"string.key\":\"dummy_value\"}}",
		},
		{
			"nested.reference.value.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/nested.ref.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/nested.ref.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"int_key" : 7,
					"bool_key": true,
					"string_key": "ref_value"
				}
			}`,
			1,
			"nested.ref.mapping",
			"{\"nested.object\":{\"bool.key\":true,\"int.key\":7,\"string.key\":\"ref_value\"}}",
		},
		{
			"json.value.reference.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/ref.path.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/ref.path.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"keys" : {
						"int" : 7,
						"bool": true,
						"string": "ref_value"
					}
				}
			}`,
			1,
			"ref.path.mapping",
			"{\"bool.key\":true,\"int.key\":7,\"string.key\":\"ref_value\"}",
		},
		{
			"missing.reference.path.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/ref.path.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/ref.path.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"int" : 7,
					"bool": true,
					"string": "ref_value"
				}
			}`,
			1,
			"ref.path.mapping",
			"{}",
		},
		{
			"field.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/field.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/field.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"string_key": "field_value"
				}
			}`,
			1,
			"field.mapping",
			"{\"string.key\":\"mapped_field_value\"}",
		},
		{
			"nested.field.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/nested.field.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/nested.field.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"string_key": "field_value"
				}
			}`,
			1,
			"nested.field.mapping",
			"{\"nested.object\":{\"string.key\":\"mapped_field_value\"}}",
		},
		{
			"default.value.field.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/field.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/field.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"string_key": "x"
				}
			}`,
			1,
			"field.mapping",
			"{\"string.key\":\"default_mapped_field_value\"}",
		},
		{
			"missing.key.field.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/field.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/field.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"int_key": "7"
				}
			}`,
			1,
			"field.mapping",
			"{}",
		},
		{
			"missing.field.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/missing.field.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/missing.field.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"int_key": "7"
				}
			}`,
			1,
			"missing.field.mapping",
			"{\"int.key\":\"7\"}",
		},
		{
			"path.field.mapping",
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/path.field.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/path.field.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"keys" : {
						"string": "field_value"
					}
				}
			}`,
			1,
			"path.field.mapping",
			"{\"string.key\":\"mapped_field_value\"}",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			d2cMessage := &routingmessage.TelemetryMessage{}
			err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
			require.NoError(t, err)

			convertedValue, ok := d2cMessage.Payload.(map[string]interface{})
			serializedValue, _ := json.Marshal(convertedValue)
			assert.True(t, ok)

			assert.Equal(t, testValues.messageType, d2cMessage.MessageType)
			assert.Equal(t, testValues.messageSubType, d2cMessage.MessageSubType)
			assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
			assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
			assert.Equal(t, testValues.convertedValue, string(serializedValue))
		})
	}
}

func TestIgnoreConvertDittoValue(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	var testData = []struct {
		jsonPayload    string
		messageSubType string
	}{
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/ignore.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/ignore.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"string_key": "field_value"
				}
			}`,
			"ignore.mapping",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/nested.ignore.mapping",
				"path": "/features/ContainerOrchestator/outbox/messages/nested.ignore.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"value": {
					"string_key": "field_value"
				}
			}`,
			"nested.ignore.mapping",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)
			assert.Nil(t, convertedMessages)
		})
	}
}

func TestErrorConvertDittoValue(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	jsonPayload := `{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/static.value.mapping",
				"headers": {
					"content-type": "application/json"
				},
				"path": "/features/ContainerOrchestator/outbox/messages/static.value.mapping",
				"value": "dummy_value"
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestSerializeJSON(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	jsonPayload := `{
			"topic": "tenant1/dummy-device:edge:containers/things/live/messages/serialize.json.object",
			"path": "/features/ContainerOrchestator/outbox/messages/serialize.json.object",
			"headers": {
				"content-type": "application/json"
			},
			"value": {
				"x": "y"
			}
	}`
	convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.NoError(t, err)

	d2cMessage := &routingmessage.TelemetryMessage{}
	err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
	require.NoError(t, err)

	convertedValue, ok := d2cMessage.Payload.(map[string]interface{})
	assert.True(t, ok)

	assert.Equal(t, 1, d2cMessage.MessageType)
	assert.Equal(t, "serialize.json.object", d2cMessage.MessageSubType)
	assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
	assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
	assert.Equal(t, map[string]interface{}{"x": "y"}, convertedValue)
}

func TestSerializeToJsonString(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	jsonPayload := `{
			"topic": "tenant1/dummy-device:edge:containers/things/live/messages/serialize.json.string",
			"path": "/features/ContainerOrchestator/outbox/messages/serialize.json.string",
			"headers": {
				"content-type": "application/json"
			},
			"value": {
				"bool_key": true,
				"int_key": 7,
				"string_key": "dummy_value"
			}
	}`
	convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.NoError(t, err)

	d2cMessage := &routingmessage.TelemetryMessage{}
	err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
	require.NoError(t, err)

	assert.Equal(t, 1, d2cMessage.MessageType)
	assert.Equal(t, "serialize.json.string", d2cMessage.MessageSubType)
	assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
	assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
	assert.Equal(t, "{\"bool_key\":true,\"int_key\":7,\"string_key\":\"dummy_value\"}", d2cMessage.Payload.(string))
}

func TestConvertDittoValueAndSerializeToBfb(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	jsonPayload := `{
			"topic": "tenant1/dummy-device:edge:containers/things/live/messages/converted.value.serialize.bfb",
			"path": "/features/ContainerOrchestator/outbox/messages/converted.value.serialize.bfb",
			"headers": {
				"content-type": "application/json"
			},
			"value": {
				"message_id": "7",
				"version": "1.0.0",
				"text": "dummy_text"
			}
	}`
	convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.NoError(t, err)

	d2cMessage := &routingmessage.TelemetryMessage{}
	err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
	require.NoError(t, err)

	convertedValue, ok := d2cMessage.Payload.(string)
	assert.True(t, ok)

	assert.Equal(t, 1, d2cMessage.MessageType)
	assert.Equal(t, "converted.value.serialize.bfb", d2cMessage.MessageSubType)
	assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
	assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
	assert.Equal(t, "CgE3EgpkdW1teV90ZXh0GgUxLjAuMA==", convertedValue)
}

func TestSerializeConvertedValueToJsonString(t *testing.T) {
	handler := createTelemetryMessageHandler(t, convertDittoValueMessageMapperConfig)
	jsonPayload := `{
			"topic": "tenant1/dummy-device:edge:containers/things/live/messages/converted.value.serialize.json.string",
			"path": "/features/ContainerOrchestator/outbox/messages/converted.value.serialize.json.string",
			"headers": {
				"content-type": "application/json"
			},
			"value": {
				"int_key" : "7",
				"bool_key": true,
				"string_key": "dummy_value"
			}
	}`
	convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.NoError(t, err)

	d2cMessage := &routingmessage.TelemetryMessage{}
	err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
	require.NoError(t, err)

	convertedValue, ok := d2cMessage.Payload.(string)
	assert.True(t, ok)

	assert.Equal(t, 1, d2cMessage.MessageType)
	assert.Equal(t, "converted.value.serialize.json.string", d2cMessage.MessageSubType)
	assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
	assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
	assert.Equal(t, "{\"bool.key\":true,\"int.key\":\"7\",\"string.key\":\"dummy_value\"}", convertedValue)
}

func TestUnsupportedMessageType(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/deleted",
		"headers": {
		  "response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/deleted",
		"value": {
		  "name": "influxdb"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestMessageTypeMappingWithoutSpecificDescriptorMappingFields(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	var testData = []struct {
		jsonPayload     string
		messageType     int
		messageSubType  string
		protobufPayload string
	}{
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noTopic",
				"headers": {
					"response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/noTopic",
				"value": {
					"message_id": "some-message-id-1234",
					"text": "simple text added",
					"version": "1.0.8"
				}
			}`,
			1,
			"no.topic",
			"ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44",
		},
		{
			`{
				"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noPath",
				"headers": {
					"response-required": false
				},
				"path": "/features/ContainerOrchestator/outbox/messages/noPath",
				"value": {
					"message_id": "some-message-id-1234",
					"text": "simple text added",
					"version": "1.0.8"
				}
			}`,
			1,
			"no.path",
			"ChRzb21lLW1lc3NhZ2UtaWQtMTIzNBIRc2ltcGxlIHRleHQgYWRkZWQaBTEuMC44",
		},
	}
	for _, testValues := range testData {
		t.Run(testValues.messageSubType, func(t *testing.T) {
			convertedMessages, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(testValues.jsonPayload)))
			require.NoError(t, err)

			d2cMessage := &routingmessage.TelemetryMessage{}
			err = json.Unmarshal(convertedMessages[0].Payload, d2cMessage)
			require.NoError(t, err)

			protobufPayload, ok := d2cMessage.Payload.(string)
			assert.True(t, ok)

			assert.Equal(t, testValues.messageType, d2cMessage.MessageType)
			assert.Equal(t, testValues.messageSubType, d2cMessage.MessageSubType)
			assert.Equal(t, "2.0", d2cMessage.EnvelopeVersion)
			assert.Equal(t, "1.0", d2cMessage.PayloadVersion)
			assert.Equal(t, testValues.protobufPayload, protobufPayload)
		})
	}
}

func TestNonExistingMessageTypeDescriptor(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noDescriptor",
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/noDescriptor",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingDittoMessageTopic(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/non.matching.topic",
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/non.matching.topic",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingDittoMessagePath(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/non.matching.path",
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/non.matching.path",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingMessageTopicKey(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/noDescriptor",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}	setUpMethod(t)
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestNonExistingMessagePath(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"topic": "tenant1/dummy-device:edge:containers/things/live/messages/noDescriptor",
		"headers": {
			"response-required": false
		},
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestUnsupportedMessagePayload(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := "topic=tenant1/dummy-device:edge:containers/things/live/messages/removed, name=influxdb"
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func TestMissingDittoTopic(t *testing.T) {
	handler := createTelemetryMessageHandler(t, commonMessageMapperConfig)
	jsonPayload := `{
		"headers": {
			"response-required": false
		},
		"path": "/features/ContainerOrchestator/outbox/messages/noDescriptor",
		"value": {
			"message_id": "some-message-id-1234",
			"text": "simple text added",
			"version": "1.0.8"
		}
	}`
	_, err := handler.HandleMessage(createWatermillMessageForD2C([]byte(jsonPayload)))
	require.Error(t, err)
}

func createTelemetryMessageHandler(t *testing.T, messageMapperConfig string) handlers.TelemetryHandler {
	mapperConfig, _ := mapperconfig.LoadMessageMapperConfig(messageMapperConfig)
	messageHandler := CreateThingsTelemetryHandler(mapperConfig, protobuf.NewProtobufJSONMarshaller(mapperConfig))
	messageHandler.Init(&config.RemoteConnectionInfo{DeviceID: "dummy-device", HubName: "dummy-hub"})
	return messageHandler
}

func createWatermillMessageForD2C(payload []byte) *message.Message {
	watermillMessage := &message.Message{
		Payload: []byte(payload),
	}
	return watermillMessage
}
