/*
	Copyright 2019 NetFoundry, Inc.

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
	"github.com/netfoundry/ziti-foundation/storage/boltz"
	"time"
)

type Entity interface {
	GetId() string
	SetId(string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetTags() map[string]interface{}
}

type BaseEntity struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Tags      map[string]interface{}
}

func (entity *BaseEntity) GetId() string {
	return entity.Id
}

func (entity *BaseEntity) SetId(id string) {
	entity.Id = id
}

func (entity *BaseEntity) GetCreatedAt() time.Time {
	return entity.CreatedAt
}

func (entity *BaseEntity) GetUpdatedAt() time.Time {
	return entity.UpdatedAt
}

func (entity *BaseEntity) GetTags() map[string]interface{} {
	return entity.Tags
}

func (entity *BaseEntity) FillCommon(boltEntity boltz.ExtEntity) {
	entity.Id = boltEntity.GetId()
	entity.CreatedAt = boltEntity.GetCreatedAt()
	entity.UpdatedAt = boltEntity.GetUpdatedAt()
	entity.Tags = boltEntity.GetTags()
}

type QueryMetaData struct {
	Count            int64
	Limit            int64
	Offset           int64
	FilterableFields []string
}
