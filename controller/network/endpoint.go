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
	"github.com/netfoundry/ziti-fabric/controller/controllers"
	"github.com/netfoundry/ziti-fabric/controller/db"
	"github.com/netfoundry/ziti-fabric/controller/models"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"reflect"
	"time"
)

type Endpoint struct {
	models.BaseEntity
	Service  string
	Router   string
	Binding  string
	Address  string
	PeerData map[uint32][]byte
}

func (entity *Endpoint) fillFrom(_ Controller, _ *bbolt.Tx, boltEntity boltz.Entity) error {
	boltEndpoint, ok := boltEntity.(*db.Endpoint)
	if !ok {
		return errors.Errorf("unexpected type %v when filling model endpoint", reflect.TypeOf(boltEntity))
	}
	entity.Service = boltEndpoint.Service
	entity.Router = boltEndpoint.Id
	entity.Binding = boltEndpoint.Binding
	entity.Address = boltEndpoint.Address
	entity.PeerData = boltEndpoint.PeerData
	entity.FillCommon(boltEndpoint)
	return nil
}

func (entity *Endpoint) toBolt() *db.Endpoint {
	return &db.Endpoint{
		BaseExtEntity: *boltz.NewExtEntity(entity.Id, entity.Tags),
		Service:       entity.Service,
		Router:        entity.Router,
		Binding:       entity.Binding,
		Address:       entity.Address,
		PeerData:      entity.PeerData,
	}
}

func newEndpointController(ctrls *Controllers) *EndpointController {
	result := &EndpointController{
		baseController: newController(ctrls, ctrls.stores.Endpoint),
		store:          ctrls.stores.Endpoint,
	}
	result.impl = result
	return result
}

type EndpointController struct {
	baseController
	store db.EndpointStore
}

func (c *EndpointController) newModelEntity() boltEntitySink {
	return &Endpoint{}
}

func (c *EndpointController) Create(s *Endpoint) (string, error) {
	var id string
	var err error
	err = c.db.Update(func(tx *bbolt.Tx) error {
		id, err = c.createInTx(boltz.NewMutateContext(tx), s)
		return err
	})
	return id, err
}

func (c *EndpointController) createInTx(ctx boltz.MutateContext, e *Endpoint) (string, error) {
	if e.Id == "" {
		e.Id = uuid.New().String()
	}
	if e.Binding == "" {
		return "", errors.Errorf("binding is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if e.Address == "" {
		return "", errors.Errorf("address is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if !c.stores.Service.IsEntityPresent(ctx.Tx(), e.Service) {
		return "", errors.Errorf("invalid service %v for new endpoint %v", e.Service, e.Id)
	}
	if e.Router == "" {
		return "", errors.Errorf("router is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if !c.stores.Router.IsEntityPresent(ctx.Tx(), e.Router) {
		return "", errors.Errorf("invalid router %v for new endpoint %v", e.Router, e.Id)
	}
	e.CreatedAt = time.Now()
	if err := c.GetStore().Create(ctx, e.toBolt()); err != nil {
		return "", err
	}
	return e.Id, nil
}

func (c *EndpointController) Update(s *Endpoint) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		return c.GetStore().Update(boltz.NewMutateContext(tx), s.toBolt(), nil)
	})
}

func (c *EndpointController) Patch(s *Endpoint, checker boltz.FieldChecker) error {
	return c.db.Update(func(tx *bbolt.Tx) error {
		return c.GetStore().Update(boltz.NewMutateContext(tx), s.toBolt(), checker)
	})
}

func (c *EndpointController) get(id string) (*Endpoint, bool) {
	var entity *Endpoint
	err := c.db.View(func(tx *bbolt.Tx) error {
		boltEntity, err := c.store.LoadOneById(tx, id)
		if err != nil {
			return err
		}
		if boltEntity == nil {
			return fmt.Errorf("missing endpoint '%s'", id)
		}
		entity = c.fromBolt(boltEntity)
		return nil
	})
	if err != nil {
		pfxlog.Logger().Errorf("failed loading endpoint (%s)", err)
		return nil, false
	}
	return entity, true
}

func (c *EndpointController) list(serviceId, routerId string) ([]*Endpoint, error) {
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
		ids, _, err := c.GetStore().QueryIds(tx, query)
		if err != nil {
			return err
		}
		for _, id := range ids {
			boltEntity, err := c.store.LoadOneById(tx, id)
			if err != nil {
				return err
			}
			entity := c.fromBolt(boltEntity)
			results = append(results, entity)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *EndpointController) Delete(id string) error {
	return controllers.DeleteEntityById(c.GetStore(), c.db, id)
}

func (c *EndpointController) fromBolt(entity *db.Endpoint) *Endpoint {
	result := &Endpoint{
		Service:  entity.Service,
		Router:   entity.Router,
		Binding:  entity.Binding,
		Address:  entity.Address,
		PeerData: entity.PeerData,
	}
	result.FillCommon(entity)
	return result
}
