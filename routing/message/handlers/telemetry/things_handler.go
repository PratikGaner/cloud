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
	"fmt"
	"strings"
	"time"

	"github.com/eclipse-kanto/suite-connector/connector"

	kantocfg "github.com/eclipse-kanto/azure-connector/config"
	"github.com/eclipse-kanto/azure-connector/routing"
	"github.com/eclipse-kanto/azure-connector/routing/message/handlers"

	routingmessage "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message"
	mapperconfig "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/config"
	"github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message/protobuf"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/pkg/errors"
)

const (
	envelopeVersion      = "2.0"
	payloadVersion       = "1.0"
	localTopics          = "event/#,e/#,telemetry/#,t/#"
	telemetryHandlerName = "things_telemetry_handler"
)

const (
	ignoreValue             = "_"
	funcTimestamp           = "timestamp()"
	fieldMappingKeyDefault  = "default"
	serializationJSONString = "jsonString"
)

type thingsTelemetryHandler struct {
	connInfo     *kantocfg.RemoteConnectionInfo
	mapperConfig *mapperconfig.MessageMapperConfig
	marshaller   protobuf.Marshaller
	incrementors map[string]int
}

// CreateThingsTelemetryHandler instantiates a things telemetry message handler
func CreateThingsTelemetryHandler(mapperConfig *mapperconfig.MessageMapperConfig, marshaller protobuf.Marshaller) handlers.TelemetryHandler {
	return &thingsTelemetryHandler{
		mapperConfig: mapperConfig,
		marshaller:   marshaller,
		incrementors: map[string]int{},
	}
}

func (h *thingsTelemetryHandler) Init(connInfo *kantocfg.RemoteConnectionInfo) error {
	h.connInfo = connInfo
	return nil
}

func (h *thingsTelemetryHandler) HandleMessage(msg *message.Message) ([]*message.Message, error) {
	dittoMessage := &protocol.Envelope{}
	if err := json.Unmarshal(msg.Payload, dittoMessage); err != nil {
		return nil, errors.Wrap(err, "cannot deserialize Ditto message!")
	}

	messageType, messageSubType, telemetryMapping, err := h.getTelemetryMapping(dittoMessage, h.mapperConfig)
	if err != nil {
		return nil, err
	}

	dittoByteValue, err := json.Marshal(dittoMessage.Value)
	if err != nil {
		return nil, errors.Wrap(err, "cannot deserialize Ditto value")
	}

	var payload interface{}
	var correlationID string

	isConverted := false
	dittoValue := dittoByteValue
	if telemetryMapping.ValueMapping != nil {
		isConverted = true
		dittoValue, correlationID, err = h.convertDittoValue(telemetryMapping, dittoValue)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("cannot convert Ditto value '%v'", dittoMessage.Value))
		}
		if dittoValue == nil {
			return nil, nil
		}
	}
	payload = dittoValue
	if telemetryMapping.ProtoFile != "" {
		payload, err = h.marshaller.Marshal(messageType, messageSubType, dittoValue)
		if err != nil {
			return nil, err
		}
	} else if telemetryMapping.Serialization == serializationJSONString {
		payload = string(dittoValue)
	} else if isConverted {
		mapValue := map[string]interface{}{}
		if err := json.Unmarshal(dittoValue, &mapValue); err != nil {
			return nil, errors.Wrap(err, "cannot serialize telemetry message payload")
		}
		payload = mapValue
	} else {
		payload = dittoMessage.Value
	}

	d2cMessage := &routingmessage.TelemetryMessage{
		MessageType:     messageType,
		MessageSubType:  messageSubType,
		Timestamp:       getUnixTimestampMs(),
		EnvelopeVersion: envelopeVersion,
		PayloadVersion:  payloadVersion,
		Payload:         payload,
	}

	if len(correlationID) == 0 {
		correlationID = dittoMessage.Headers.CorrelationID()
	}
	d2cMessage.CorrelationID = correlationID

	outgoingPayload, err := json.Marshal(d2cMessage)
	if err != nil {
		return nil, errors.Wrap(err, "cannot serialize D2C message")
	}

	msgID := watermill.NewUUID()
	outgoingMessage := message.NewMessage(msgID, outgoingPayload)
	outgoingTopic := routing.CreateTelemetryTopic(h.connInfo.DeviceID, msgID)
	outgoingMessage.SetContext(connector.SetTopicToCtx(outgoingMessage.Context(), outgoingTopic))
	return []*message.Message{outgoingMessage}, nil
}

func (h *thingsTelemetryHandler) getTelemetryMapping(dittoMessage *protocol.Envelope, mapperConfig *mapperconfig.MessageMapperConfig) (int, string, *mapperconfig.TelemetryMessageMapping, error) {
	if dittoMessage.Topic == nil {
		return -1, "", nil, errors.New("missing Ditto topic in message")
	}

	topic := dittoMessage.Topic.String()
	path := dittoMessage.Path
	telemetryMappings, err := mapperConfig.GetTelemetryMessageMappings()
	if err != nil {
		return -1, "", nil, errors.Wrap(err, fmt.Sprintf("cannot map Ditto topic '%s' & Ditto path '%s' to D2C message sub type", topic, path))
	}

	for messageType, telemetryMessageTypeMappings := range telemetryMappings {
		for messageSubType, messageMapping := range telemetryMessageTypeMappings {
			mappingProperties := messageMapping.MappingProperties
			if mappingProperties.Topic != "" {
				if mappingProperties.Path != "" {
					if strings.Contains(topic, mappingProperties.Topic) && strings.Contains(path, mappingProperties.Path) {
						return messageType, messageSubType, messageMapping, nil
					}
				} else {
					if strings.Contains(topic, mappingProperties.Topic) {
						return messageType, messageSubType, messageMapping, nil
					}
				}
			} else if mappingProperties.Path != "" {
				if strings.Contains(path, mappingProperties.Path) {
					return messageType, messageSubType, messageMapping, nil
				}
			}
		}
	}
	return -1, "", nil, fmt.Errorf("cannot map Ditto topic '%s' & Ditto path '%s' to D2C message sub type", topic, path)
}

func (h *thingsTelemetryHandler) convertDittoValue(telemetryMapping *mapperconfig.TelemetryMessageMapping, dittoValue []byte) ([]byte, string, error) {
	var err error
	valueMap := map[string]interface{}{}
	if err = json.Unmarshal(dittoValue, &valueMap); err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("cannot deserialize Ditto value '%v'", dittoValue))
	}
	correlationID := ""
	if cID, ok := valueMap["correlationId"]; ok {
		correlationID = cID.(string)
	}
	valueMapping := deepCopyMap(telemetryMapping.ValueMapping)
	if ok := h.convertDittoValueInternal(telemetryMapping, valueMapping, valueMap); !ok {
		return nil, correlationID, nil
	}
	convertedValue, err := json.Marshal(valueMapping)
	return convertedValue, correlationID, err
}

func (h *thingsTelemetryHandler) convertDittoValueInternal(telemetryMapping *mapperconfig.TelemetryMessageMapping, valueMapping, valueMap map[string]interface{}) bool {
	for key, value := range valueMapping {
		switch convertedValue := value.(type) {
		case string:
			if strings.HasPrefix(convertedValue, "$") {
				refPath := strings.Split(convertedValue[1:], ".")
				refValue := h.getRefValue(refPath, valueMap)
				if refValue == nil {
					delete(valueMapping, key)
					continue
				}
				fieldValue := h.getFieldMappingValue(telemetryMapping, convertedValue, refValue)
				if fieldValue == ignoreValue {
					return false
				}
				valueMapping[key] = fieldValue
			} else if convertedValue == funcTimestamp {
				valueMapping[key] = getUnixTimestampMs()
			} else if strings.HasPrefix(convertedValue, "++") {
				incrementKey := convertedValue[2:]
				increment := h.incrementors[incrementKey] + 1
				h.incrementors[incrementKey] = increment
				valueMapping[key] = increment
			} else {
				valueMapping[key] = convertedValue
			}
		case map[string]interface{}:
			if ok := h.convertDittoValueInternal(telemetryMapping, convertedValue, valueMap); !ok {
				return false
			}
		}
	}
	return true
}

func (h *thingsTelemetryHandler) getFieldMappingValue(telemetryMapping *mapperconfig.TelemetryMessageMapping, fieldKey string, value interface{}) interface{} {
	if telemetryMapping.FieldMappings == nil {
		return value
	}
	fieldMapping := telemetryMapping.FieldMappings[fieldKey]
	if fieldMapping == nil {
		return value
	}
	for fieldKey, fieldValue := range fieldMapping {
		if fieldKey == value {
			return fieldValue
		}
	}
	return fieldMapping[fieldMappingKeyDefault]
}

func (h *thingsTelemetryHandler) getRefValue(path []string, value map[string]interface{}) interface{} {
	return h.getRefValueInternal(path, 0, value)
}

func (h *thingsTelemetryHandler) getRefValueInternal(path []string, index int, value map[string]interface{}) interface{} {
	if index == len(path)-1 {
		return value[path[index]]
	}
	var ok bool
	var valueMap map[string]interface{}
	if valueMap, ok = value[path[index]].(map[string]interface{}); !ok {
		return nil
	}
	return h.getRefValueInternal(path, index+1, valueMap)
}

func deepCopyMap(originMap map[string]interface{}) map[string]interface{} {
	copyMap := make(map[string]interface{})
	for key, value := range originMap {
		switch valueType := value.(type) {
		case map[string]interface{}:
			copyMap[key] = deepCopyMap(valueType)
		default:
			copyMap[key] = value
		}
	}
	return copyMap
}

func getUnixTimestampMs() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func (h *thingsTelemetryHandler) Name() string {
	return telemetryHandlerName
}

func (h *thingsTelemetryHandler) Topics() string {
	return localTopics
}
