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

package db

import (
	"github.com/netfoundry/ziti-foundation/storage/ast"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
)

import (
	"fmt"
	"go.etcd.io/bbolt"
)

const (
	EntityTypeRouters      = "routers"
	FieldRouterFingerprint = "fingerprint"
)

type Router struct {
	Id          string
	Fingerprint string
}

func (entity *Router) GetId() string {
	return entity.Id
}

func (entity *Router) SetId(id string) {
	entity.Id = id
}

func (entity *Router) LoadValues(_ boltz.CrudStore, bucket *boltz.TypedBucket) {
	entity.Fingerprint = bucket.GetStringOrError(FieldRouterFingerprint)
}

func (entity *Router) SetValues(ctx *boltz.PersistContext) {
	ctx.SetString(FieldRouterFingerprint, entity.Fingerprint)
}

func (entity *Router) GetEntityType() string {
	return EntityTypeRouters
}

type RouterStore interface {
	boltz.CrudStore
	LoadOneById(tx *bbolt.Tx, id string) (*Router, error)
}

func newRouterStore(stores *stores) *routerStoreImpl {
	notFoundErrorFactory := func(id string) error {
		return fmt.Errorf("missing router '%s'", id)
	}

	store := &routerStoreImpl{
		baseStore: baseStore{
			stores:    stores,
			BaseStore: boltz.NewBaseStore(nil, EntityTypeRouters, notFoundErrorFactory, boltz.RootBucket),
		},
	}
	store.InitImpl(store)
	store.AddSymbol(FieldRouterFingerprint, ast.NodeTypeString)
	return store
}

type routerStoreImpl struct {
	baseStore
}

func (store *routerStoreImpl) initializeLinked() {
	store.AddFkSetSymbol(EntityTypeEndpoints, store.stores.endpoint)
}

func (store *routerStoreImpl) NewStoreEntity() boltz.BaseEntity {
	return &Router{}
}

func (store *routerStoreImpl) LoadOneById(tx *bbolt.Tx, id string) (*Router, error) {
	entity := &Router{}
	if found, err := store.BaseLoadOneById(tx, id, entity); !found || err != nil {
		return nil, err
	}
	return entity, nil
}

func (store *routerStoreImpl) DeleteById(ctx boltz.MutateContext, id string) error {
	endpointIds := store.GetRelatedEntitiesIdList(ctx.Tx(), id, EntityTypeEndpoints)
	for _, endpointId := range endpointIds {
		if err := store.stores.endpoint.DeleteById(ctx, endpointId); err != nil {
			return err
		}
	}
	return store.BaseStore.DeleteById(ctx, id)
}
