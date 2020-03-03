/*
	Copyright 2020 NetFoundry, Inc.

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

	https://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package handler_mgmt

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/controller/handler_common"
	"github.com/netfoundry/ziti-fabric/controller/network"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
	"reflect"
)

type listEndpointsHandler struct {
	network *network.Network
}

func newListEndpointsHandler(network *network.Network) *listEndpointsHandler {
	return &listEndpointsHandler{network: network}
}

func (h *listEndpointsHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_ListEndpointsRequestType)
}

func (h *listEndpointsHandler) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	ls := &mgmt_pb.ListEndpointsRequest{}
	if err := proto.Unmarshal(msg.Body, ls); err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}
	response := &mgmt_pb.ListEndpointsResponse{Endpoints: make([]*mgmt_pb.Endpoint, 0)}

	result, err := h.network.Endpoints.BaseList(ls.Query)
	if err == nil {
		for _, entity := range result.Entities {
			endpoint, ok := entity.(*network.Endpoint)
			if !ok {
				errorMsg := fmt.Sprintf("unexpected result in endpoint list of type: %v", reflect.TypeOf(entity))
				handler_common.SendFailure(msg, ch, errorMsg)
				return
			}
			response.Endpoints = append(response.Endpoints, toApiEndpoint(endpoint))
		}

		body, err := proto.Marshal(response)
		if err == nil {
			responseMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_ListEndpointsResponseType), body)
			responseMsg.ReplyTo(msg)
			if err := ch.Send(responseMsg); err != nil {
				pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error sending response (%s)", err)
			}
		} else {
			handler_common.SendFailure(msg, ch, err.Error())
		}
	} else {
		handler_common.SendFailure(msg, ch, err.Error())
	}
}
