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

import "github.com/netfoundry/ziti-fabric/controller/db"

type env struct {
	db        *db.Db
	stores    *db.Stores
	endpoints *endpointController
	routers   *routerController
	services  *serviceController
}

func (e *env) getDb() *db.Db {
	return e.db
}

func initEnv(db *db.Db, stores *db.Stores) *env {
	result := &env{
		db:     db,
		stores: stores,
	}
	result.endpoints = newEndpointController(result)
	result.routers = newRouterController(result)
	result.services = newServiceController(result)
	return result
}
