package network

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/openziti/fabric/controller/db"
	"github.com/openziti/fabric/controller/xt"
	"github.com/openziti/fabric/controller/xt_smartrouting"
	"github.com/openziti/fabric/ctrl_msg"
	"testing"
)

func TestRouteSender_DestroysTerminatorWhenInvalidOnHandleRouteSendAndWeControl(t *testing.T) {
	ctx := db.NewTestContext(t)
	defer ctx.Cleanup()

	closeNotify := make(chan struct{})
	defer close(closeNotify)

	network, err := NewNetwork("test", nil, ctx.GetDb(), nil, NewVersionProviderTest(), closeNotify)
	ctx.NoError(err)

	entityHelper := newTestEntityHelper(ctx, network)
	logger := pfxlog.ChannelLogger("test")

	router1 := entityHelper.addTestRouter()
	router2 := entityHelper.addTestRouter()
	path := &Path{
		Nodes: []*Router{router1, router2},
	}

	svc := entityHelper.addTestService("svc")

	identity := "identity"

	term := entityHelper.addTestTerminator(svc.Id, router1.Id, identity, true)
	term.Binding = "edge"

	network.Terminators.Create(term)

	errCode := byte(ctrl_msg.ErrorTypeInvalidTerminator)

	rs := routeSender{
		serviceCounters: network,
		terminators:     network.Terminators,
		attendance:      make(map[string]bool),
	}

	status := &RouteStatus{
		Router:    router1,
		ErrorCode: &errCode,
		Success:   true,
		Attempt:   1,
	}

	peerData, cleanup, err := rs.handleRouteSend(1, path, xt_smartrouting.NewFactory().NewStrategy(), status, term, logger)
	ctx.NoError(err)
	ctx.Nil(peerData)
	ctx.Nil(cleanup)

	newTerm, err := network.Terminators.Read(term.Id)
	ctx.Error(err)
	ctx.Nil(newTerm)
}

func TestRouteSender_SetPrecidenceToNilTerminatorWhenInvalidOnHandleRouteSendAndWeDontControl(t *testing.T) {
	ctx := db.NewTestContext(t)
	defer ctx.Cleanup()

	closeNotify := make(chan struct{})
	defer close(closeNotify)

	network, err := NewNetwork("test", nil, ctx.GetDb(), nil, NewVersionProviderTest(), closeNotify)
	ctx.NoError(err)

	entityHelper := newTestEntityHelper(ctx, network)
	logger := pfxlog.ChannelLogger("test")

	router1 := entityHelper.addTestRouter()
	router2 := entityHelper.addTestRouter()
	path := &Path{
		Nodes: []*Router{router1, router2},
	}

	svc := entityHelper.addTestService("svc")

	identity := "identity"

	term := entityHelper.addTestTerminator(svc.Id, router1.Id, identity, true)
	term.Binding = "DNE"

	network.Terminators.Create(term)

	errCode := byte(ctrl_msg.ErrorTypeInvalidTerminator)

	rs := routeSender{
		serviceCounters: network,
		terminators:     network.Terminators,
		attendance:      make(map[string]bool),
	}

	status := &RouteStatus{
		Router:    router1,
		ErrorCode: &errCode,
		Success:   true,
		Attempt:   1,
	}

	peerData, cleanup, err := rs.handleRouteSend(1, path, xt_smartrouting.NewFactory().NewStrategy(), status, term, logger)
	ctx.NoError(err)
	ctx.Nil(peerData)
	ctx.Nil(cleanup)

	newTerm, err := network.Terminators.Read(term.Id)
	ctx.NoError(err)
	ctx.NotNil(newTerm)

	ctx.Equal(xt.Precedences.Failed, newTerm.GetPrecedence())
}
