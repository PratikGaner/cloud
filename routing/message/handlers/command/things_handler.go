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
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"

	"github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	mapperconfig "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/config"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/protobuf"

	"github.com/eclipse-kanto/suite-connector/connector"

	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/pkg/errors"
)

const (
	commandThingsHandlerName     = "command_things_handler"
	dittoNamespace               = "azure.edge"
	messageTopicPattern          = "command///req/%s/%s"
	messageTopicPatternWithThing = "command//%s:%s:%s/req/%s/%s"
	dittoTopicPatternWithThing   = `"%s/%s:%s/things/live/messages/%s"`
)

const (
	keyCorrelationID = "correlationId"
	keyPayload       = "payload"
)

type thingsCommandHandler struct {
	connInfo     *config.RemoteConnectionInfo
	mapperConfig *mapperconfig.MessageMapperConfig
	marshaller   protobuf.Marshaller
}

// CreateThingsCommandHandler instantiates a things command message handler
func CreateThingsCommandHandler(mapperConfig *mapperconfig.MessageMapperConfig, marshaller protobuf.Marshaller) handlers.CommandHandler {
	return &thingsCommandHandler{
		mapperConfig: mapperConfig,
		marshaller:   marshaller,
	}
}

func (h *thingsCommandHandler) Init(deviceInfo *config.RemoteConnectionInfo) error {
	h.connInfo = deviceInfo
	return nil
}

func (h *thingsCommandHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	cloudMessage, err := parseCommandMessage(msg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot deserialize cloud message")
	}
	messageMapping, err := h.mapperConfig.GetCommandMessageMapping(cloudMessage.CommandName)
	if err != nil {
		return nil, err
	}
	mappingProperties := messageMapping.MappingProperties

	deviceID := h.connInfo.HubName + ":" + h.connInfo.DeviceID
	topicStr := fmt.Sprintf(dittoTopicPatternWithThing, dittoNamespace, deviceID, mappingProperties.Thing, mappingProperties.Action)
	topic := &protocol.Topic{}
	if err := topic.UnmarshalJSON([]byte(topicStr)); err != nil {
		return nil, err
	}

	headers := protocol.NewHeaders(protocol.WithContentType("application/json"), protocol.WithCorrelationID(cloudMessage.CorrelationID))

	var dittoValue interface{}
	if messageMapping.ProtoFile == "" {
		wrappedPayload := wrapDittoPayload(mappingProperties, cloudMessage.Payload)
		if messageMapping.RetainCorrelationID {
			dittoValue = map[string]interface{}{
				keyCorrelationID: cloudMessage.CorrelationID,
				keyPayload:       wrappedPayload,
			}
		} else {
			dittoValue = wrappedPayload
		}
	} else {
		var bytePayload []byte
		bytePayload, err = h.marshaller.Unmarshal(cloudMessage.CommandName, cloudMessage.Payload.(string))
		if err == nil {
			mapValue := map[string]interface{}{}
			if err = json.Unmarshal(bytePayload, &mapValue); err == nil {
				dittoValue = mapValue
			}
		}
	}
	if err != nil {
		return nil, err
	}

	dittoMessage := &protocol.Envelope{
		Topic:   topic,
		Headers: headers,
		Path:    mappingProperties.Path,
		Value:   dittoValue,
	}

	outgoingPayload, err := json.Marshal(dittoMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot serialize C2D message")
	}
	outgoingMessage := message.NewMessage(watermill.NewUUID(), outgoingPayload)
	outgoingTopic := createMessageTopic(mappingProperties, deviceID, cloudMessage.CorrelationID)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), outgoingTopic))

	return []*message.Message{outgoingMessage}, nil
}

func createMessageTopic(mappingProperties *mapperconfig.CommandMappingProperties, deviceID, reqID string) string {
	if mappingProperties.Thing == "" {
		return fmt.Sprintf(messageTopicPattern, reqID, mappingProperties.Action)
	}
	return fmt.Sprintf(messageTopicPatternWithThing, dittoNamespace, deviceID, mappingProperties.Thing, reqID, mappingProperties.Action)
}

func wrapDittoPayload(mappingProperties *mapperconfig.CommandMappingProperties, payload interface{}) interface{} {
	if mappingProperties.Value != "" {
		wrappedPayload := make(map[string]interface{})

		var jsonMap map[string]interface{}
		if json.Unmarshal([]byte(fmt.Sprint(payload)), &jsonMap) != nil {
			wrappedPayload[mappingProperties.Value] = payload
		} else {
			wrappedPayload[mappingProperties.Value] = jsonMap
		}

		return wrappedPayload
	}
	return payload
}

func (h *thingsCommandHandler) Name() string {
	return commandThingsHandlerName
}
