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
	"github.com/netfoundry/ziti-fabric/controller/controllers"
	"github.com/netfoundry/ziti-fabric/controller/db"
	"github.com/netfoundry/ziti-fabric/controller/models"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"github.com/orcaman/concurrent-map"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"reflect"
)

type Service struct {
	models.BaseEntity
	EndpointStrategy string
	Endpoints        []*Endpoint
}

func (entity *Service) fillFrom(ctrl Controller, tx *bbolt.Tx, boltEntity boltz.Entity) error {
	boltEndpoint, ok := boltEntity.(*db.Service)
	if !ok {
		return errors.Errorf("unexpected type %v when filling model endpoint", reflect.TypeOf(boltEntity))
	}
	entity.EndpointStrategy = boltEndpoint.EndpointStrategy
	entity.FillCommon(boltEndpoint)

	endpointIds := ctrl.getControllers().stores.Service.GetRelatedEntitiesIdList(tx, entity.Id, db.EntityTypeEndpoints)
	for _, endpointId := range endpointIds {
		if endpoint, _ := ctrl.getControllers().Endpoints.readInTx(tx, endpointId); endpoint != nil {
			entity.Endpoints = append(entity.Endpoints, endpoint)
		}
	}

	return nil
}

func (entity *Service) toBolt() *db.Service {
	return &db.Service{
		BaseExtEntity:    *boltz.NewExtEntity(entity.Id, entity.Tags),
		EndpointStrategy: entity.EndpointStrategy,
	}
}

type ServiceController struct {
	baseController
	cache cmap.ConcurrentMap
	store db.ServiceStore
}

func newServiceController(controllers *Controllers) *ServiceController {
	result := &ServiceController{
		baseController: newController(controllers, controllers.stores.Service),
		cache:          cmap.New(),
		store:          controllers.stores.Service,
	}

	controllers.stores.Endpoint.On(boltz.EventCreate, result.endpointChanged)
	controllers.stores.Endpoint.On(boltz.EventUpdate, result.endpointChanged)
	controllers.stores.Endpoint.On(boltz.EventDelete, result.endpointChanged)

	return result
}

func (ctrl *ServiceController) endpointChanged(params ...interface{}) {
	if len(params) > 0 {
		if entity, ok := params[0].(*db.Endpoint); ok {
			ctrl.RemoveFromCache(entity.Service)
		}
	}
}

func (ctrl *ServiceController) Create(s *Service) error {
	err := ctrl.db.Update(func(tx *bbolt.Tx) error {
		ctx := boltz.NewMutateContext(tx)
		if err := ctrl.store.Create(ctx, s.toBolt()); err != nil {
			return err
		}
		for _, endpoint := range s.Endpoints {
			endpoint.Service = s.Id
			if _, err := ctrl.Endpoints.createInTx(ctx, endpoint); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	ctrl.cache.Set(s.Id, s)
	return nil
}

func (ctrl *ServiceController) Update(s *Service) error {
	err := ctrl.db.Update(func(tx *bbolt.Tx) error {
		return ctrl.store.Update(boltz.NewMutateContext(tx), s.toBolt(), nil)
	})
	if err != nil {
		return err
	}
	ctrl.cache.Set(s.Id, s)
	return nil
}

func (ctrl *ServiceController) Read(id string) (entity *Service, err error) {
	err = ctrl.db.View(func(tx *bbolt.Tx) error {
		entity, err = ctrl.readInTx(tx, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	return entity, err
}

func (ctrl *ServiceController) readInTx(tx *bbolt.Tx, id string) (*Service, error) {
	if t, found := ctrl.cache.Get(id); found {
		return t.(*Service), nil
	}

	entity := &Service{}
	if err := ctrl.readEntityInTx(tx, id, entity); err != nil {
		return nil, err
	}

	ctrl.cache.Set(id, entity)
	return entity, nil
}

func (ctrl *ServiceController) Delete(id string) error {
	err := controllers.DeleteEntityById(ctrl.store, ctrl.db, id)
	ctrl.cache.Remove(id)
	return err
}

func (ctrl *ServiceController) RemoveFromCache(id string) {
	ctrl.cache.Remove(id)
}
