package xt_default

import (
	"github.com/netfoundry/ziti-fabric/controller/xt"
	"github.com/pkg/errors"
)

const (
	Name = "single-endpoint"
)

func NewFactory() xt.Factory {
	return factory{}
}

type factory struct{}

func (f factory) GetStrategyName() string {
	return Name
}

func (f factory) GetStrategyAliases() []string {
	return []string{"", "default"}
}

func (f factory) NewStrategy() xt.Strategy {
	return strategy{}
}

type strategy struct{}

func (s strategy) NotifyEvent(event xt.TerminatorEvent) {
	// nothing to do. We're only ever going to have one, so changing precedence or weights doesn't affect anything
}

func (s strategy) Select(terminators []xt.WeightedTerminator, _ uint32) (xt.Terminator, error) {
	return terminators[0], nil
}

func (s strategy) HandleTerminatorChange(event xt.StrategyChangeEvent) error {
	terminatorCount := len(event.GetCurrent()) + len(event.GetAdded())
	if terminatorCount > 1 {
		return errors.Errorf("strategy %v only allows a single terminator. service %v has %v",
			Name, event.GetServiceId(), terminatorCount)
	}
	return nil
}
