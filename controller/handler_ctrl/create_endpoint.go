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

package handler_ctrl

import (
	"github.com/golang/protobuf/proto"
	"github.com/netfoundry/ziti-fabric/controller/handler_common"
	"github.com/netfoundry/ziti-fabric/controller/network"
	"github.com/netfoundry/ziti-fabric/pb/ctrl_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
)

type createEndpointHandler struct {
	router  *network.Router
	network *network.Network
}

func newCreateEndpointHandler(network *network.Network, router *network.Router) *createEndpointHandler {
	return &createEndpointHandler{
		network: network,
		router:  router,
	}
}

func (h *createEndpointHandler) ContentType() int32 {
	return int32(ctrl_pb.ContentType_CreateEndpointRequestType)
}

func (h *createEndpointHandler) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	request := &ctrl_pb.CreateEndpointRequest{}
	if err := proto.Unmarshal(msg.Body, request); err != nil {
		handler_common.SendFailure(msg, ch, err.Error())
		return
	}
	endpoint := &network.Endpoint{
		BaseEntity: network.BaseEntity{
			Id: request.Id,
		},
		Service:  request.ServiceId,
		Router:   h.router.Id,
		Binding:  request.Binding,
		Address:  request.Address,
		PeerData: request.PeerData,
	}

	if err := h.network.CreateEndpoint(endpoint); err == nil {
		handler_common.SendSuccess(msg, ch, "")
	} else {
		handler_common.SendFailure(msg, ch, err.Error())
	}
}
