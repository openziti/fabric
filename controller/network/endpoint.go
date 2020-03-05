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
	"fmt"
	"github.com/google/uuid"
	"github.com/michaelquigley/pfxlog"
	"github.com/netfoundry/ziti-fabric/controller/db"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"time"
)

type Endpoint struct {
	Id        string
	Service   string
	Router    *Router
	Binding   string
	Address   string
	CreatedAt time.Time
	PeerData  map[uint32][]byte
}

func (entity *Endpoint) toBolt() *db.Endpoint {
	return &db.Endpoint{
		Id:       entity.Id,
		Service:  entity.Service,
		Router:   entity.Router.Id,
		Binding:  entity.Binding,
		Address:  entity.Address,
		PeerData: entity.PeerData,
	}
}

type endpointController struct {
	*env
	store db.EndpointStore
}

func newEndpointController(env *env) *endpointController {
	return &endpointController{
		env:   env,
		store: env.stores.Endpoint,
	}
}

func (c *endpointController) create(s *Endpoint) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		return c.createInTx(boltz.NewMutateContext(tx), s)
	})
}

func (c *endpointController) createInTx(ctx boltz.MutateContext, e *Endpoint) error {
	if e.Id == "" {
		e.Id = uuid.New().String()
	}
	if e.Binding == "" {
		return errors.Errorf("binding is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if e.Address == "" {
		return errors.Errorf("address is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if !c.stores.Service.IsEntityPresent(ctx.Tx(), e.Service) {
		return errors.Errorf("invalid service %v for new endpoint %v", e.Service, e.Id)
	}
	if e.Router == nil {
		return errors.Errorf("router is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if !c.stores.Router.IsEntityPresent(ctx.Tx(), e.Router.Id) {
		return errors.Errorf("invalid router %v for new endpoint %v", e.Router.Id, e.Id)
	}
	e.CreatedAt = time.Now()
	return c.store.Create(ctx, e.toBolt())
}

func (c *endpointController) update(s *Endpoint) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		return c.store.Update(boltz.NewMutateContext(tx), s.toBolt(), nil)
	})
}

func (c *endpointController) get(id string) (*Endpoint, bool) {
	var entity *Endpoint
	err := c.db.View(func(tx *bbolt.Tx) error {
		boltEntity, err := c.store.LoadOneById(tx, id)
		if err != nil {
			return err
		}
		if boltEntity == nil {
			return fmt.Errorf("missing endpoint '%s'", id)
		}
		entity, err = c.fromBolt(boltEntity)
		return err
	})
	if err != nil {
		pfxlog.Logger().Errorf("failed loading endpoint (%s)", err)
		return nil, false
	}
	return entity, true
}

func (c *endpointController) list(serviceId, routerId string) ([]*Endpoint, error) {
	var results []*Endpoint
	err := c.db.View(func(tx *bbolt.Tx) error {
		query := ""
		if serviceId != "" {
			query = fmt.Sprintf(`service = "%v"`, serviceId)
		}
		if routerId != "" {
			if query != "" {
				query += " and "
			}
			query += fmt.Sprintf(`router = "%v"`, routerId)
		}
		if query == "" {
			query = "true"
		}
		ids, _, err := c.store.QueryIds(tx, query)
		if err != nil {
			return err
		}
		for _, id := range ids {
			boltEntity, err := c.store.LoadOneById(tx, id)
			if err != nil {
				return err
			}
			entity, err := c.fromBolt(boltEntity)
			if err != nil {
				return err
			}
			results = append(results, entity)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *endpointController) remove(id string) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		return c.store.DeleteById(boltz.NewMutateContext(tx), id)
	})
}

func (c *endpointController) fromBolt(entity *db.Endpoint) (*Endpoint, error) {
	router, err := c.routers.get(entity.Router)
	if err != nil {
		return nil, err
	}
	return &Endpoint{
		Id:       entity.Id,
		Service:  entity.Service,
		Router:   router,
		Binding:  entity.Binding,
		Address:  entity.Address,
		PeerData: entity.PeerData,
	}, nil
}
