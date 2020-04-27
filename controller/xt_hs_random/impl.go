package xt_hs_random

import (
	"github.com/netfoundry/ziti-fabric/controller/xt"
	"math/rand"
)

func NewFactory() xt.Factory {
	return factory{}
}

type factory struct{}

func (f factory) GetStrategyName() string {
	return "hs-random"
}

func (f factory) GetStrategyAliases() []string {
	return nil
}

func (f factory) NewStrategy() xt.Strategy {
	return strategy{}
}

type strategy struct{}

func (s strategy) Select(terminators []xt.WeightedTerminator, totalWeight uint32) (xt.Terminator, error) {
	count := len(terminators)
	if count == 1 {
		return terminators[0], nil
	}
	selected := rand.Intn(count)
	return terminators[selected], nil
}

func (s strategy) NotifyEvent(xt.TerminatorEvent) {}

func (s strategy) HandleTerminatorChange(xt.StrategyChangeEvent) error {
	return nil
}
