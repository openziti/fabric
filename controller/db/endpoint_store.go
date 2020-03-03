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
	"encoding/binary"
	"fmt"
	"github.com/netfoundry/ziti-foundation/storage/ast"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"go.etcd.io/bbolt"
	"time"
)

const (
	EntityTypeEndpoints    = "endpoints"
	FieldEndpointService   = "service"
	FieldEndpointRouter    = "router"
	FieldEndpointBinding   = "binding"
	FieldEndpointAddress   = "address"
	FieldEndpointCreatedAt = "createdAt"
	FieldServerPeerData    = "peerdata"
	// FieldEndpointStrategyInputs = "strategyInputs"
)

type Endpoint struct {
	Id        string
	Service   string
	Router    string
	Binding   string
	Address   string
	CreatedAt time.Time
	PeerData  map[uint32][]byte
	// StrategyInputs interface{}
}

func (entity *Endpoint) GetId() string {
	return entity.Id
}

func (entity *Endpoint) SetId(id string) {
	entity.Id = id
}

func (entity *Endpoint) LoadValues(_ boltz.CrudStore, bucket *boltz.TypedBucket) {
	entity.Service = bucket.GetStringOrError(FieldEndpointService)
	entity.Router = bucket.GetStringOrError(FieldEndpointRouter)
	entity.Binding = bucket.GetStringOrError(FieldEndpointBinding)
	entity.Address = bucket.GetStringWithDefault(FieldEndpointAddress, "")
	entity.CreatedAt = bucket.GetTimeOrError(FieldEndpointCreatedAt)

	data := bucket.GetBucket(FieldServerPeerData)
	if data != nil {
		entity.PeerData = make(map[uint32][]byte)
		iter := data.Cursor()
		for k, v := iter.First(); k != nil; k, v = iter.Next() {
			entity.PeerData[binary.LittleEndian.Uint32(k)] = v
		}
	}
}

func (entity *Endpoint) SetValues(ctx *boltz.PersistContext) {
	ctx.SetString(FieldEndpointService, entity.Service)
	ctx.SetString(FieldEndpointRouter, entity.Router)
	ctx.SetString(FieldEndpointBinding, entity.Binding)
	ctx.SetString(FieldEndpointAddress, entity.Address)
	ctx.SetTimeP(FieldEndpointCreatedAt, &entity.CreatedAt)

	_ = ctx.Bucket.DeleteBucket([]byte(FieldServerPeerData))
	if entity.PeerData != nil {
		hostDataBucket := ctx.Bucket.GetOrCreateBucket(FieldServerPeerData)
		for k, v := range entity.PeerData {
			key := make([]byte, 4)
			binary.LittleEndian.PutUint32(key, k)
			hostDataBucket.PutValue(key, v)
		}
	}
}

func (entity *Endpoint) GetEntityType() string {
	return EntityTypeEndpoints
}

type EndpointStore interface {
	boltz.CrudStore
	LoadOneById(tx *bbolt.Tx, id string) (*Endpoint, error)
}

func newEndpointStore(stores *stores) *endpointStoreImpl {
	notFoundErrorFactory := func(id string) error {
		return fmt.Errorf("missing endpoint '%s'", id)
	}

	store := &endpointStoreImpl{
		baseStore: baseStore{
			stores:    stores,
			BaseStore: boltz.NewBaseStore(nil, EntityTypeServices, notFoundErrorFactory, boltz.RootBucket),
		},
	}
	store.InitImpl(store)
	store.AddSymbol(FieldEndpointBinding, ast.NodeTypeString)
	store.AddSymbol(FieldEndpointAddress, ast.NodeTypeString)
	store.AddSymbol(FieldEndpointCreatedAt, ast.NodeTypeDatetime)
	return store
}

type endpointStoreImpl struct {
	baseStore
}

func (store *endpointStoreImpl) NewStoreEntity() boltz.BaseEntity {
	return &Endpoint{}
}

func (store *endpointStoreImpl) initializeLinked() {
	store.AddFkSymbol(FieldEndpointService, store.stores.service)
	store.AddFkSymbol(FieldEndpointRouter, store.stores.router)
}

func (store *endpointStoreImpl) LoadOneById(tx *bbolt.Tx, id string) (*Endpoint, error) {
	entity := &Endpoint{}
	if found, err := store.BaseLoadOneById(tx, id, entity); !found || err != nil {
		return nil, err
	}
	return entity, nil
}
