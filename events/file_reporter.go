package events

import (
	"github.com/openziti/fabric/controller/network"
	"github.com/openziti/foundation/events"
	"github.com/openziti/foundation/identity/identity"
	"github.com/openziti/foundation/metrics/metrics_pb"
)

func registerJsonLoggerEventHandlerType(config map[interface{}]interface{}) (EventHandlerFactory, bool)  {

	rep := &FabricHandler{
		Config: config,
	}

	return rep, true
}




func (handler *FabricHandler) NewEventHandler(config map[interface{}]interface{}) (interface{}, error) {

	// this is a bit weird, didn't know a better way to pass the config down
	handler.Emitter.SetConfig(config)

	return &FabricHandler{
		name: "FabricHandler",
		Config: config,
	}, nil

}

type FabricHandler struct {
	name   string
	Config map[interface{}]interface{}
	Emitter events.GlobalHandler
}

func (handler *FabricHandler) SessionCreated(sessionId *identity.TokenId, clientId *identity.TokenId, serviceId string, circuit *network.Circuit) {

	message := &events.SessionMessage{
		Namespace: "fabric.sessions",
		EventType: "created",
		SessionId: sessionId.Token,
		ClientId: clientId.Token,
		ServiceId: serviceId,
		Circuit: circuit.String(),
	}

	handler.Emitter.Emit(message)
}

func (handler *FabricHandler) SessionDeleted(sessionId *identity.TokenId, clientId *identity.TokenId) {

	message := &events.SessionMessage{
		Namespace: "fabric.sessions",
		EventType: "deleted",
		SessionId: sessionId.Token,
		ClientId: clientId.Token,
	}

	handler.Emitter.Emit(message)

}

func (handler *FabricHandler) CircuitUpdated(sessionId *identity.TokenId, circuit *network.Circuit) {

	message := &events.SessionMessage{
		Namespace: "fabric.sessions",
		EventType: "circuitUpdated",
		SessionId: sessionId.Token,
		Circuit: circuit.String(),
	}

	handler.Emitter.Emit(message)
}

func (handler *FabricHandler) AcceptMetrics(message *metrics_pb.MetricsMessage) {
	//logger := pfxlog.Logger()
	//logger.Info("MSG: %v", message)

	// handler.Emit(message)

}
