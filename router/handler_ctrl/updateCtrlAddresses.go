package handler_ctrl

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/channel/v2"
	"github.com/openziti/fabric/pb/ctrl_pb"
	"google.golang.org/protobuf/proto"
)

type CtrlAddressUpdater interface {
	UpdateCtrlEndpoints(endpoints []string) error
}

type updateCtrlAddressesHandler struct {
	CtrlAddressUpdater
}

func (handler *updateCtrlAddressesHandler) ContentType() int32 {
	return int32(ctrl_pb.ContentType_UpdateCtrlAddressesType)
}

func (handler *updateCtrlAddressesHandler) HandleReceive(msg *channel.Message, ch channel.Channel) {
	upd := &ctrl_pb.UpdateCtrlAddresses{}
	if err := proto.Unmarshal(msg.Body, upd); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("error unmarshalling")
		return
	}

	if err := handler.UpdateCtrlEndpoints(upd.Addresses); err != nil {
		pfxlog.ContextLogger(ch.Label()).WithError(err).Error("Unable to update ctrl addresses")
	}
}

func newUpdateCtrlAddressesHandler(callback CtrlAddressUpdater) channel.TypedReceiveHandler {
	return &updateCtrlAddressesHandler{
		callback,
	}
}
