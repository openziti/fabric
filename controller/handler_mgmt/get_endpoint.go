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
	"github.com/golang/protobuf/ptypes"
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/controller/handler_common"
	"github.com/netfoundry/ziti-fabric/controller/network"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
)

type getEndpointHandler struct {
	network *network.Network
}

func newGetEndpointHandler(network *network.Network) *getEndpointHandler {
	return &getEndpointHandler{network: network}
}

func (h *getEndpointHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_GetEndpointRequestType)
}

func (h *getEndpointHandler) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	rs := &mgmt_pb.GetEndpointRequest{}
	if err := proto.Unmarshal(msg.Body, rs); err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}
	response := &mgmt_pb.GetEndpointResponse{}
	endpoint, err := h.network.Endpoints.Read(rs.EndpointId)
	if err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}

	response.Endpoint = toApiEndpoint(endpoint)
	body, err := proto.Marshal(response)
	if err == nil {
		responseMsg := channel2.NewMessage(int32(mgmt_pb.ContentType_GetEndpointResponseType), body)
		responseMsg.ReplyTo(msg)
		ch.Send(responseMsg)

	} else {
		pfxlog.ContextLogger(ch.Label()).Errorf("unexpected error (%s)", err)
	}
}

func toApiEndpoint(s *network.Endpoint) *mgmt_pb.Endpoint {
	ts, err := ptypes.TimestampProto(s.CreatedAt)
	if err != nil {
		pfxlog.Logger().Warnf("unexpected bad timestamp conversion: %v", err)
	}
	return &mgmt_pb.Endpoint{
		Id:        s.Id,
		ServiceId: s.Service,
		RouterId:  s.Router,
		Binding:   s.Binding,
		Address:   s.Address,
		CreatedAt: ts,
	}
}
