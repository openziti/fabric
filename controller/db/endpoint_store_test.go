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
	"fmt"
	"github.com/google/uuid"
	"github.com/netfoundry/ziti-foundation/util/stringz"
	"go.etcd.io/bbolt"
	"testing"
	"time"
)

func Test_EndpointStore(t *testing.T) {
	ctx := NewTestContext(t)
	defer ctx.Cleanup()
	ctx.Init()

	t.Run("test create invalid endpoints", ctx.testCreateInvalidEndpoints)
	t.Run("test create/delete endpoints", ctx.testCreateEndpoints)
	t.Run("test create/delete endpoints", ctx.testLoadQueryEndpoints)
	t.Run("test update Services", ctx.testUpdateEndpoints)
	t.Run("test delete Services", ctx.testDeleteEndpoints)
}

func (ctx *TestContext) testCreateInvalidEndpoints(t *testing.T) {
	ctx.NextTest(t)
	defer ctx.cleanupAll()

	endpoint := &Endpoint{}
	err := ctx.Create(endpoint)
	ctx.EqualError(err, "cannot create endpoint with blank id")

	service := ctx.requireNewService()
	router := ctx.requireNewRouter()

	endpoint.Id = uuid.New().String()
	endpoint.Service = uuid.New().String()
	endpoint.Router = router.Id
	err = ctx.Create(endpoint)
	ctx.EqualError(err, fmt.Sprintf("no entity of type services with id %v", endpoint.Service))

	endpoint.Id = uuid.New().String()
	endpoint.Service = service.Id
	endpoint.Router = uuid.New().String()
	err = ctx.Create(endpoint)
	ctx.EqualError(err, fmt.Sprintf("no entity of type routers with id %v", endpoint.Router))
}

type endpointTestEntities struct {
	service  *Service
	service2 *Service

	router  *Router
	router2 *Router

	endpoint  *Endpoint
	endpoint2 *Endpoint
	endpoint3 *Endpoint
}

func (ctx *TestContext) createTestEndpoints() *endpointTestEntities {
	e := &endpointTestEntities{}

	e.service = ctx.requireNewService()
	e.router = ctx.requireNewRouter()

	e.endpoint = &Endpoint{}
	e.endpoint.Id = uuid.New().String()
	e.endpoint.Service = e.service.Id
	e.endpoint.Router = e.router.Id
	e.endpoint.Binding = uuid.New().String()
	e.endpoint.Address = uuid.New().String()
	ctx.RequireCreate(e.endpoint)

	e.router2 = ctx.requireNewRouter()

	e.endpoint2 = &Endpoint{}
	e.endpoint2.Id = uuid.New().String()
	e.endpoint2.Service = e.service.Id
	e.endpoint2.Router = e.router2.Id
	ctx.RequireCreate(e.endpoint2)

	e.service2 = ctx.requireNewService()

	e.endpoint3 = &Endpoint{}
	e.endpoint3.Id = uuid.New().String()
	e.endpoint3.Service = e.service2.Id
	e.endpoint3.Router = e.router2.Id
	ctx.RequireCreate(e.endpoint3)

	return e
}

func (ctx *TestContext) testCreateEndpoints(t *testing.T) {
	ctx.NextTest(t)
	defer ctx.cleanupAll()

	e := ctx.createTestEndpoints()

	ctx.ValidateBaseline(e.endpoint)
	ctx.ValidateBaseline(e.endpoint2)
	ctx.ValidateBaseline(e.endpoint3)

	endpointIds := ctx.GetRelatedIds(e.service, EntityTypeEndpoints)
	ctx.EqualValues(2, len(endpointIds))
	ctx.True(stringz.Contains(endpointIds, e.endpoint.Id))
	ctx.True(stringz.Contains(endpointIds, e.endpoint2.Id))

	endpointIds = ctx.GetRelatedIds(e.router, EntityTypeEndpoints)
	ctx.EqualValues(1, len(endpointIds))
	ctx.EqualValues(e.endpoint.Id, endpointIds[0])

	endpointIds = ctx.GetRelatedIds(e.router2, EntityTypeEndpoints)
	ctx.EqualValues(2, len(endpointIds))
	ctx.True(stringz.Contains(endpointIds, e.endpoint2.Id))
	ctx.True(stringz.Contains(endpointIds, e.endpoint3.Id))

	endpointIds = ctx.GetRelatedIds(e.service2, EntityTypeEndpoints)
	ctx.EqualValues(1, len(endpointIds))
	ctx.EqualValues(e.endpoint3.Id, endpointIds[0])

}

func (ctx *TestContext) testLoadQueryEndpoints(t *testing.T) {
	ctx.NextTest(t)
	defer ctx.cleanupAll()

	e := ctx.createTestEndpoints()

	err := ctx.GetDb().View(func(tx *bbolt.Tx) error {
		loadedEndpoint, err := ctx.stores.Endpoint.LoadOneById(tx, e.endpoint.Id)
		ctx.NoError(err)
		ctx.NotNil(loadedEndpoint)
		ctx.EqualValues(e.endpoint.Id, loadedEndpoint.Id)
		ctx.EqualValues(e.endpoint.Service, loadedEndpoint.Service)
		ctx.EqualValues(e.endpoint.Router, loadedEndpoint.Router)
		ctx.EqualValues(e.endpoint.Binding, loadedEndpoint.Binding)
		ctx.EqualValues(e.endpoint.Address, loadedEndpoint.Address)

		ids, _, err := ctx.stores.Endpoint.QueryIdsf(tx, `service = "%v"`, e.service.Id)
		ctx.NoError(err)
		ctx.EqualValues(2, len(ids))
		ctx.True(stringz.Contains(ids, e.endpoint.Id))
		ctx.True(stringz.Contains(ids, e.endpoint2.Id))

		ids, _, err = ctx.stores.Endpoint.QueryIdsf(tx, `router = "%v"`, e.router2.Id)
		ctx.NoError(err)
		ctx.EqualValues(2, len(ids))
		ctx.True(stringz.Contains(ids, e.endpoint2.Id))
		ctx.True(stringz.Contains(ids, e.endpoint3.Id))

		ids, _, err = ctx.stores.Service.QueryIdsf(tx, `anyOf(endpoints) = "%v"`, e.endpoint.Id)
		ctx.NoError(err)
		ctx.EqualValues(1, len(ids))
		ctx.True(stringz.Contains(ids, e.service.Id))

		ids, _, err = ctx.stores.Router.QueryIdsf(tx, `anyOf(endpoints) = "%v"`, e.endpoint.Id)
		ctx.NoError(err)
		ctx.EqualValues(1, len(ids))
		ctx.True(stringz.Contains(ids, e.router.Id))

		return nil
	})
	ctx.NoError(err)
}

func (ctx *TestContext) testUpdateEndpoints(t *testing.T) {
	ctx.NextTest(t)
	defer ctx.cleanupAll()

	e := ctx.createTestEndpoints()

	endpoint := e.endpoint
	ctx.RequireReload(endpoint)

	time.Sleep(time.Millisecond * 10) // ensure updatedAt is after createdAt

	endpoint.Service = e.service2.Id
	endpoint.Router = e.router2.Id
	endpoint.Binding = uuid.New().String()
	endpoint.Address = uuid.New().String()
	endpoint.Tags = ctx.CreateTags()
	ctx.RequireUpdate(endpoint)
	ctx.ValidateUpdated(endpoint)
}

func (ctx *TestContext) testDeleteEndpoints(t *testing.T) {
	ctx.NextTest(t)
	defer ctx.cleanupAll()

	e := ctx.createTestEndpoints()

	ctx.RequireDelete(e.endpoint3)
	ctx.RequireDelete(e.router2)

	ctx.ValidateDeleted(e.endpoint2.Id)
	ctx.ValidateDeleted(e.endpoint3.Id)

	ctx.RequireDelete(e.service)
	ctx.ValidateDeleted(e.endpoint.Id)
}
