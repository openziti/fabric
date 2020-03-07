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
	"github.com/netfoundry/ziti-fabric/controller/db"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"go.etcd.io/bbolt"
)

type Controllers struct {
	db        *db.Db
	stores    *db.Stores
	Endpoints *EndpointController
	Routers   *RouterController
	Services  *ServiceController
}

func (e *Controllers) getDb() *db.Db {
	return e.db
}

func NewControllers(db *db.Db, stores *db.Stores) *Controllers {
	result := &Controllers{
		db:     db,
		stores: stores,
	}
	result.Endpoints = newEndpointController(result)
	result.Routers = newRouterController(result)
	result.Services = newServiceController(result)
	return result
}

type Controller interface {
	BaseLoad(id string) (Entity, error)

	getStore() boltz.CrudStore
	newModelEntity() boltEntitySink
	readEntityInTx(tx *bbolt.Tx, id string, modelEntity boltEntitySink) error
}

type boltEntitySink interface {
	Entity
	fillFrom(controller Controller, tx *bbolt.Tx, boltEntity boltz.Entity) error
}

type baseController struct {
	*Controllers
	impl Controller
}

func (ctrl *baseController) BaseLoad(id string) (Entity, error) {
	entity := ctrl.impl.newModelEntity()
	if err := ctrl.readEntity(id, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (ctrl *baseController) readEntity(id string, modelEntity boltEntitySink) error {
	return ctrl.db.View(func(tx *bbolt.Tx) error {
		return ctrl.readEntityInTx(tx, id, modelEntity)
	})
}

func (ctrl *baseController) readEntityInTx(tx *bbolt.Tx, id string, modelEntity boltEntitySink) error {
	boltEntity := ctrl.impl.getStore().NewStoreEntity()
	found, err := ctrl.impl.getStore().BaseLoadOneById(tx, id, boltEntity)
	if err != nil {
		return err
	}
	if !found {
		return boltz.NewNotFoundError(ctrl.impl.getStore().GetSingularEntityType(), "id", id)
	}

	return modelEntity.fillFrom(ctrl.impl, tx, boltEntity)
}
