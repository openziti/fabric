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
	"github.com/netfoundry/ziti-fabric/controller/network"
	"github.com/netfoundry/ziti-fabric/pb/mgmt_pb"
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/pkg/errors"
	"time"
)

type createEndpointHandler struct {
	network *network.Network
}

func newCreateEndpointHandler(network *network.Network) *createEndpointHandler {
	return &createEndpointHandler{network: network}
}

func (h *createEndpointHandler) ContentType() int32 {
	return int32(mgmt_pb.ContentType_CreateEndpointRequestType)
}

func (h *createEndpointHandler) HandleReceive(msg *channel2.Message, ch channel2.Channel) {
	cs := &mgmt_pb.CreateEndpointRequest{}
	if err := proto.Unmarshal(msg.Body, cs); err != nil {
		sendFailure(msg, ch, err.Error())
		return
	}
	endpoint, err := toModelEndpoint(h.network, cs.Endpoint)
	if err != nil {
		sendFailure(msg, ch, err.Error())
		return
	}
	if err = h.network.CreateEndpoint(endpoint); err == nil {
		sendSuccess(msg, ch, "")
	} else {
		sendFailure(msg, ch, err.Error())
	}
}

func toModelEndpoint(n *network.Network, e *mgmt_pb.Endpoint) (*network.Endpoint, error) {
	router, _ := n.GetRouter(e.RouterId)
	if router == nil {
		return nil, errors.Errorf("invalid router id %v", e.RouterId)
	}

	binding := "transport"
	if e.Binding != "" {
		binding = e.Binding
	}

	return &network.Endpoint{
		Id:        e.Id,
		Service:   e.ServiceId,
		Router:    router,
		Binding:   binding,
		Address:   e.Address,
		CreatedAt: time.Now(),
		PeerData:  nil,
	}, nil
}
