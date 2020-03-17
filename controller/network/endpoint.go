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
	"github.com/google/uuid"
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
	entity.Router = boltEndpoint.Router
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

func newEndpointController(controllers *Controllers) *EndpointController {
	result := &EndpointController{
		baseController: newController(controllers, controllers.stores.Endpoint),
		store:          controllers.stores.Endpoint,
	}
	result.impl = result
	return result
}

type EndpointController struct {
	baseController
	store db.EndpointStore
}

func (ctrl *EndpointController) newModelEntity() boltEntitySink {
	return &Endpoint{}
}

func (ctrl *EndpointController) Create(s *Endpoint) (string, error) {
	var id string
	var err error
	err = ctrl.db.Update(func(tx *bbolt.Tx) error {
		id, err = ctrl.createInTx(boltz.NewMutateContext(tx), s)
		return err
	})
	return id, err
}

func (ctrl *EndpointController) createInTx(ctx boltz.MutateContext, e *Endpoint) (string, error) {
	if e.Id == "" {
		e.Id = uuid.New().String()
	}
	if e.Binding == "" {
		return "", models.NewFieldError("required value is missing", "binding", e.Binding)
	}
	if e.Address == "" {
		return "", models.NewFieldError("required value is missing", "address", e.Binding)
	}
	if !ctrl.stores.Service.IsEntityPresent(ctx.Tx(), e.Service) {
		return "", boltz.NewNotFoundError("service", "service", e.Service)
	}
	if e.Router == "" {
		return "", errors.Errorf("router is required when creating endpoint. id: %v, service: %v", e.Id, e.Service)
	}
	if !ctrl.stores.Router.IsEntityPresent(ctx.Tx(), e.Router) {
		return "", boltz.NewNotFoundError("router", "router", e.Router)
	}
	e.CreatedAt = time.Now()
	if err := ctrl.GetStore().Create(ctx, e.toBolt()); err != nil {
		return "", err
	}
	return e.Id, nil
}

func (ctrl *EndpointController) Update(s *Endpoint) error {
	return ctrl.db.Update(func(tx *bbolt.Tx) error {
		return ctrl.GetStore().Update(boltz.NewMutateContext(tx), s.toBolt(), nil)
	})
}

func (ctrl *EndpointController) Patch(s *Endpoint, checker boltz.FieldChecker) error {
	return ctrl.db.Update(func(tx *bbolt.Tx) error {
		return ctrl.GetStore().Update(boltz.NewMutateContext(tx), s.toBolt(), checker)
	})
}

func (ctrl *EndpointController) Read(id string) (entity *Endpoint, err error) {
	err = ctrl.db.View(func(tx *bbolt.Tx) error {
		entity, err = ctrl.readInTx(tx, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return entity, err
}

func (ctrl *EndpointController) readInTx(tx *bbolt.Tx, id string) (*Endpoint, error) {
	entity := &Endpoint{}
	err := ctrl.readEntityInTx(tx, id, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (ctrl *EndpointController) Delete(id string) error {
	return controllers.DeleteEntityById(ctrl.GetStore(), ctrl.db, id)
}
