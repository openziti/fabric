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
	"github.com/netfoundry/ziti-fabric/controller/controllers"
	"github.com/netfoundry/ziti-fabric/controller/db"
	"github.com/netfoundry/ziti-foundation/channel2"
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"github.com/netfoundry/ziti-foundation/transport"
	"github.com/netfoundry/ziti-foundation/util/concurrenz"
	"github.com/orcaman/concurrent-map"
	"go.etcd.io/bbolt"
)

type Router struct {
	Id                 string
	Fingerprint        string
	AdvertisedListener transport.Address
	Control            channel2.Channel
	CostFactor         int
	Connected          concurrenz.AtomicBoolean
}

func (entity *Router) toBolt() *db.Router {
	return &db.Router{
		Id:          entity.Id,
		Fingerprint: entity.Fingerprint,
	}
}

func NewRouter(id, fingerprint string) *Router {
	return &Router{
		Id:          id,
		Fingerprint: fingerprint,
	}
}

func newRouter(id string, fingerprint string, advLstnr transport.Address, ctrl channel2.Channel) *Router {
	return &Router{
		Id:                 id,
		Fingerprint:        fingerprint,
		AdvertisedListener: advLstnr,
		Control:            ctrl,
		CostFactor:         1,
	}
}

type RouterController struct {
	*Controllers
	cache     cmap.ConcurrentMap
	connected cmap.ConcurrentMap
	store     db.RouterStore
}

func newRouterController(env *Controllers) *RouterController {
	return &RouterController{
		Controllers: env,
		cache:       cmap.New(),
		connected:   cmap.New(),
		store:       env.stores.Router,
	}
}

func (c *RouterController) markConnected(r *Router) {
	r.Connected.Set(true)
	c.connected.Set(r.Id, r)
}

func (c *RouterController) markDisconnected(r *Router) {
	r.Connected.Set(false)
	c.connected.Remove(r.Id)
}

func (c *RouterController) isConnected(id string) bool {
	return c.connected.Has(id)
}

func (c *RouterController) getConnected(id string) *Router {
	if t, found := c.connected.Get(id); found {
		return t.(*Router)
	}
	return nil
}

func (c *RouterController) allConnected() []*Router {
	var routers []*Router
	for i := range c.connected.IterBuffered() {
		routers = append(routers, i.Val.(*Router))
	}
	return routers
}

func (c *RouterController) connectedCount() int {
	return c.connected.Count()
}

func (c *RouterController) Create(router *Router) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		return c.store.Create(boltz.NewMutateContext(tx), router.toBolt())
	})
	if err != nil {
		c.cache.Set(router.Id, router)
	}
	return err
}

func (c *RouterController) Delete(id string) error {
	err := controllers.DeleteEntityById(c.store, c.db, id)
	c.cache.Remove(id)
	return err
}

func (c *RouterController) get(id string) (*Router, error) {
	if t, found := c.cache.Get(id); found {
		return t.(*Router), nil
	}

	var router *Router
	err := c.db.View(func(tx *bbolt.Tx) error {
		boltRouter, err := c.store.LoadOneById(tx, id)
		if err != nil {
			return err
		}
		if boltRouter == nil {
			return fmt.Errorf("missing router '%s'", id)
		}
		router = c.fromBolt(boltRouter)
		return nil
	})
	if err != nil {
		return nil, err
	}
	c.cache.Set(id, router)
	return router, err
}

func (c *RouterController) all() ([]*Router, error) {
	var routers []*Router
	err := c.db.View(func(tx *bbolt.Tx) error {
		ids, _, err := c.store.QueryIds(tx, "true")
		if err != nil {
			return err
		}
		for _, id := range ids {
			router, err := c.store.LoadOneById(tx, id)
			if err != nil {
				return err
			}
			routers = append(routers, c.fromBolt(router))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return routers, nil
}

func (c *RouterController) fromBolt(entity *db.Router) *Router {
	return &Router{
		Id:          entity.Id,
		Fingerprint: entity.Fingerprint,
	}
}
