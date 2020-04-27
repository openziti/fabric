package xt

import (
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"sync"
	"sync/atomic"
)

func init() {
	globalRegistry = &defaultRegistry{
		factories: &copyOnWriteFactoryMap{
			value: &atomic.Value{},
			lock:  &sync.Mutex{},
		},
		strategies: &copyOnWriteStrategyMap{
			value: &atomic.Value{},
			lock:  &sync.Mutex{},
		},
		lock: &sync.Mutex{},
	}

	globalRegistry.factories.value.Store(map[string]Factory{})
	globalRegistry.strategies.value.Store(map[string]Strategy{})
}

func GlobalRegistry() Registry {
	return globalRegistry
}

var globalRegistry *defaultRegistry

type defaultRegistry struct {
	factories  *copyOnWriteFactoryMap
	strategies *copyOnWriteStrategyMap
	lock       *sync.Mutex
}

func (registry *defaultRegistry) RegisterFactory(factory Factory) {
	registry.factories.put(factory.GetStrategyName(), factory)
	for _, alias := range factory.GetStrategyAliases() {
		registry.factories.put(alias, factory)
	}
}

func (registry *defaultRegistry) GetStrategy(name string) (Strategy, error) {
	result := registry.strategies.get(name)
	if result == nil {
		registry.lock.Lock()
		defer registry.lock.Unlock()
		result = registry.strategies.get(name)
		if result != nil {
			return result, nil
		}

		factory := registry.factories.get(name)
		if factory == nil {
			return nil, boltz.NewNotFoundError("terminatorStrategy", "name", name)
		}

		result = factory.NewStrategy()
		registry.strategies.put(factory.GetStrategyName(), result)
		for _, alias := range factory.GetStrategyAliases() {
			registry.strategies.put(alias, result)
		}
	}

	return result, nil
}

type copyOnWriteFactoryMap struct {
	value *atomic.Value
	lock  *sync.Mutex
}

func (m *copyOnWriteFactoryMap) put(key string, value Factory) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var current = m.value.Load().(map[string]Factory)
	mapCopy := map[string]Factory{}
	for k, v := range current {
		mapCopy[k] = v
	}
	mapCopy[key] = value
	m.value.Store(mapCopy)
}

func (m *copyOnWriteFactoryMap) get(key string) Factory {
	var current = m.value.Load().(map[string]Factory)
	return current[key]
}

type copyOnWriteStrategyMap struct {
	value *atomic.Value
	lock  *sync.Mutex
}

func (m *copyOnWriteStrategyMap) put(key string, value Strategy) {
	m.lock.Lock()
	defer m.lock.Unlock()

	var current = m.value.Load().(map[string]Strategy)
	mapCopy := map[string]Strategy{}
	for k, v := range current {
		mapCopy[k] = v
	}
	mapCopy[key] = value
	m.value.Store(mapCopy)
}

func (m *copyOnWriteStrategyMap) get(key string) Strategy {
	var current = m.value.Load().(map[string]Strategy)
	return current[key]
}
