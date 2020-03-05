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
	"github.com/golang/protobuf/proto"
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/controller/network"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
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
		sendFailure(msg, ch, err.Error())
		return
	}
	response := &mgmt_pb.ListEndpointsResponse{Endpoints: make([]*mgmt_pb.Endpoint, 0)}

	svcs, err := h.network.ListEndpoints(ls.ServiceId, ls.RouterId)
	if err == nil {
		for _, s := range svcs {
			response.Endpoints = append(response.Endpoints, toApiEndpoint(s))
		}

		body, err := proto.Marshal(response)
		if err == nil {
			responseMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_ListEndpointsResponseType), body)
			responseMsg.ReplyTo(msg)
			if err := ch.Send(responseMsg); err != nil {
				pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error sending response (%s)", err)
			}
		} else {
			sendFailure(msg, ch, err.Error())
		}
	} else {
		sendFailure(msg, ch, err.Error())
	}
}
