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

package network

import (
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/controller/controllers"
	"github.com/netfoundry/ziti-fabric/controller/db"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

type Service struct {
	Id               string
	EndpointStrategy string
	Endpoints        []*Endpoint
}

func (entity *Service) toBolt() *db.Service {
	return &db.Service{
		Id:               entity.Id,
		EndpointStrategy: entity.EndpointStrategy,
	}
}

type ServiceController struct {
	*Controllers
	cache cmap.ConcurrentMap
	store db.ServiceStore
}

func newServiceController(env *Controllers) *ServiceController {
	result := &ServiceController{
		Controllers: env,
		cache:       cmap.New(),
		store:       env.stores.Service,
	}

	env.stores.Endpoint.On(boltz.EventCreate, result.endpointChanged)
	env.stores.Endpoint.On(boltz.EventUpdate, result.endpointChanged)
	env.stores.Endpoint.On(boltz.EventDelete, result.endpointChanged)

	return result
}

func (c *ServiceController) endpointChanged(params ...interface{}) {
	if len(params) > 0 {
		if entity, ok := params[0].(*db.Endpoint); ok {
			c.RemoveFromCache(entity.Service)
		}
	}
}

func (c *ServiceController) Create(s *Service) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		ctx := boltz.NewMutateContext(tx)
		if err := c.store.Create(ctx, s.toBolt()); err != nil {
			return err
		}
		for _, endpoint := range s.Endpoints {
			endpoint.Service = s.Id
			if _, err := c.Endpoints.createInTx(ctx, endpoint); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	c.cache.Set(s.Id, s)
	return nil
}

func (c *ServiceController) Update(s *Service) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		return c.store.Update(boltz.NewMutateContext(tx), s.toBolt(), nil)
	})
	if err != nil {
		return err
	}
	c.cache.Set(s.Id, s)
	return nil
}

func (c *ServiceController) get(id string) *Service {
	if t, found := c.cache.Get(id); found {
		return t.(*Service)
	}
	var svc *Service
	err := c.db.View(func(tx *bbolt.Tx) error {
		svc = c.loadInTx(tx, id)
		return nil
	})
	if err != nil {
		pfxlog.Logger().Errorf("failed loading service (%s)", err)
		return nil
	}
	return svc
}

func (c *ServiceController) getInTx(tx *bbolt.Tx, id string) *Service {
	if t, found := c.cache.Get(id); found {
		return t.(*Service)
	}
	return c.loadInTx(tx, id)
}

func (c *ServiceController) loadInTx(tx *bbolt.Tx, id string) *Service {
	boltSvc, err := c.store.LoadOneById(tx, id)
	if err != nil || boltSvc == nil {
		return nil
	}
	svc := c.fromBolt(boltSvc)

	endpointIds := c.store.GetRelatedEntitiesIdList(tx, id, db.EntityTypeEndpoints)
	for _, endpointId := range endpointIds {
		if endpoint, _ := c.Endpoints.get(endpointId); endpoint != nil {
			svc.Endpoints = append(svc.Endpoints, endpoint)
		}
	}

	c.cache.Set(svc.Id, svc)
	return svc
}

func (c *ServiceController) all() ([]*Service, error) {
	var services []*Service
	err := c.db.View(func(tx *bbolt.Tx) error {
		ids, _, err := c.store.QueryIds(tx, "true")
		if err != nil {
			return err
		}
		for _, id := range ids {
			service := c.getInTx(tx, id)
			if service == nil {
				return errors.Errorf("unable to load service with id %v", id)
			}
			services = append(services, service)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return services, nil
}

func (c *ServiceController) Delete(id string) error {
	err := controllers.DeleteEntityById(c.store, c.db, id)
	c.cache.Remove(id)
	return err
}

func (c *ServiceController) RemoveFromCache(id string) {
	c.cache.Remove(id)
}

func (c *ServiceController) fromBolt(entity *db.Service) *Service {
	return &Service{
		Id:               entity.Id,
		EndpointStrategy: entity.EndpointStrategy,
	}
}
