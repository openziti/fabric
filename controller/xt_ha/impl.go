package xt_ha

import (
	"github.com/netfoundry/ziti-fabric/controller/xt"
	"sync/atomic"
)

func NewFactory() xt.Factory {
	return factory{}
}

type factory struct{}

func (f factory) GetStrategyName() string {
	return "ha"
}

func (f factory) GetStrategyAliases() []string {
	return nil
}

func (f factory) NewStrategy() xt.Strategy {
	return strategy{}
}

type strategy struct {
	xt.DefaultEventVisitor
	failCount int32
}

func (s strategy) VisitDialFailed(event xt.TerminatorEvent) {
	failCount := atomic.AddInt32(&s.failCount, 1)
	if failCount >= 3 {
		xt.GlobalCosts().SetPrecedence(event.GetTerminator().GetId(), xt.Precedences.Failed)
	}
}

func (s strategy) VisitDialSucceeded(event xt.TerminatorEvent) {
	atomic.StoreInt32(&s.failCount, 0)
}

func (s strategy) Select(terminators []xt.WeightedTerminator, _ uint32) (xt.Terminator, error) {
	return terminators[0], nil
}

func (s strategy) NotifyEvent(event xt.TerminatorEvent) {
	event.Accept(s)
}

func (s strategy) HandleTerminatorChange(xt.StrategyChangeEvent) error {
	return nil
}
