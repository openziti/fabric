package xt_hs_weighted

import (
	"github.com/netfoundry/ziti-fabric/controller/xt"
	"math"
)

func NewFactory() xt.Factory {
	return factory{}
}

type factory struct{}

func (f factory) GetStrategyName() string {
	return "hs-smartrouting"
}

func (f factory) GetStrategyAliases() []string {
	return nil
}

func (f factory) NewStrategy() xt.Strategy {
	return strategy{}
}

type strategy struct {
	xt.DefaultEventVisitor
}

func (s strategy) Select(terminators []xt.WeightedTerminator, totalWeight uint32) (xt.Terminator, error) {
	return terminators[0], nil
}

func (s strategy) NotifyEvent(event xt.TerminatorEvent) {
	event.Accept(s)
}

func (s strategy) VisitDialFailed(event xt.TerminatorEvent) {
	costs := xt.GlobalCosts()
	cost := costs.GetPrecedenceCost(event.GetTerminator().GetId())
	if cost > 0 {
		nextCost := int(cost) + 20
		if nextCost < 0 {
			nextCost = 0
		}
		costs.SetPrecedenceCost(event.GetTerminator().GetId(), uint8(nextCost))
	}
}

func (s strategy) VisitDialSucceeded(event xt.TerminatorEvent) {
	costs := xt.GlobalCosts()
	cost := costs.GetPrecedenceCost(event.GetTerminator().GetId())
	if cost < math.MaxUint8 {
		costs.SetPrecedenceCost(event.GetTerminator().GetId(), cost+1)
	}
}

func (s strategy) VisitSessionEnded(event xt.TerminatorEvent) {
	costs := xt.GlobalCosts()
	cost := costs.GetPrecedenceCost(event.GetTerminator().GetId())
	if cost < math.MaxUint8 {
		costs.SetPrecedenceCost(event.GetTerminator().GetId(), cost-2)
	}
}

func (s strategy) HandleTerminatorChange(xt.StrategyChangeEvent) error {
	return nil
}
