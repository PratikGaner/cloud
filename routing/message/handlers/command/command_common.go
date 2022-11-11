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
	"context"
	"encoding/json"

	routingmessage "github.com/eclipse-leda/leda-contrib-cloud-connector/routing/message"

	"github.com/ThreeDotsLabs/watermill/message"
)

type contextKey int

const (
	commandMessageContextKey contextKey = 4 + iota //the rest of the context keys are defined in the connector package
)

func parseCommandMessage(msg *message.Message) (*routingmessage.CloudMessage, error) {
	value, ok := msg.Context().Value(commandMessageContextKey).(*routingmessage.CloudMessage)
	if ok {
		return value, nil
	}
	cloudMessage := &routingmessage.CloudMessage{}
	if err := json.Unmarshal(msg.Payload, cloudMessage); err != nil {
		return nil, err
	}
	msg.SetContext(context.WithValue(msg.Context(), commandMessageContextKey, cloudMessage))
	return cloudMessage, nil
}
