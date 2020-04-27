package xt_hs_weighted

import (
	"github.com/netfoundry/ziti-fabric/controller/xt"
	"math"
	"math/rand"
)

func NewFactory() xt.Factory {
	return factory{}
}

type factory struct{}

func (f factory) GetStrategyName() string {
	return "hs-weighted"
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
	if len(terminators) == 1 {
		return terminators[0], nil
	}
	selected := uint32(rand.Int31n(int32(totalWeight)))
	currentWeight := uint32(0)
	for _, terminator := range terminators {
		currentWeight += terminator.GetRouteWeight()
		if selected <= currentWeight {
			return terminator, nil
		}
	}
	return terminators[0], nil
}

func (s strategy) NotifyEvent(event xt.TerminatorEvent) {
	event.Accept(s)
}

func (s strategy) VisitDialFailed(event xt.TerminatorEvent) {
	weights := xt.GlobalCosts()
	weight := weights.GetPrecedenceCost(event.GetTerminator().GetId())
	if weight > 0 {
		nextWeight := int(weight) + 20
		if nextWeight < 0 {
			nextWeight = 0
		}
		weights.SetPrecedenceCost(event.GetTerminator().GetId(), uint8(nextWeight))
	}
}

func (s strategy) VisitDialSucceeded(event xt.TerminatorEvent) {
	weights := xt.GlobalCosts()
	weight := weights.GetPrecedenceCost(event.GetTerminator().GetId())
	if weight < math.MaxUint8 {
		weights.SetPrecedenceCost(event.GetTerminator().GetId(), weight-1)
	}
}

func (s strategy) HandleTerminatorChange(xt.StrategyChangeEvent) error {
	return nil
}
