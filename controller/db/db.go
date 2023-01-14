/*
	Copyright NetFoundry Inc.

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
	"github.com/openziti/storage/boltz"
	"go.etcd.io/bbolt"
)

const (
	RootBucket   = "ziti"
	RaftIndexKey = "raftIndex"
)

type Db struct {
	boltz.Db
}

func (self *Db) UpdateWithIndex(index uint64, f func(tx *bbolt.Tx) error) error {
	if index == 0 {
		return self.Update(f)
	}

	return self.Update(func(tx *bbolt.Tx) error {
		if err := f(tx); err != nil {
			return err
		}
		return SetRaftIndex(tx, index)
	})
}

func GetRaftIndex(tx *bbolt.Tx) uint64 {
	v, found := GetMetadata(tx, func(b *boltz.TypedBucket) (uint64, bool) {
		if val := b.GetInt64(RaftIndexKey); val != nil {
			return uint64(*val), true
		}
		return 0, false
	})

	if found {
		return v
	}

	return 0
}

func Open(path string) (boltz.Db, error) {
	db, err := boltz.Open(path, RootBucket)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(RootBucket))
		return err
	})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func SetMetadata(tx *bbolt.Tx, f func(b *boltz.TypedBucket)) error {
	b := boltz.GetOrCreatePath(tx, boltz.Metadata)
	f(b)
	if b.HasError() {
		return b.GetError()
	}
	return nil
}

func GetMetadata[T any](tx *bbolt.Tx, f func(b *boltz.TypedBucket) (T, bool)) (T, bool) {
	var result T
	if b := boltz.Path(tx, boltz.Metadata); b != nil {
		return f(b)
	}
	return result, false
}

func SetRaftIndex(tx *bbolt.Tx, val uint64) error {
	return SetMetadata(tx, func(b *boltz.TypedBucket) {
		b.SetInt64(RaftIndexKey, int64(val), nil)
	})
}
