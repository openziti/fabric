/*
	(c) Copyright NetFoundry, Inc.

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

package xlink_transwarp

import (
	"github.com/emirpasic/gods/trees/btree"
	"github.com/emirpasic/gods/utils"
	"sync"
)

type txBuffer struct {
	tree      *btree.Tree
	highWater int32
	capacity  int
	lock      *sync.Mutex
	ready     *sync.Cond
}

func newTxBuffer() *txBuffer {
	t := &txBuffer{
		tree:      btree.NewWith(10240, utils.Int32Comparator),
		highWater: -1,
		capacity:  8,
		lock:      new(sync.Mutex),
	}
	t.ready = sync.NewCond(t.lock)
	return t
}

func (self *txBuffer) tx(m *message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for self.capacity < 1 {
		self.ready.Wait()
	}

	self.tree.Put(m.sequence, m)
	self.capacity--
	if m.sequence > self.highWater {
		self.highWater = m.sequence
	}
}

func (self *txBuffer) release(upTo int32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for _, sequence := range self.tree.Keys() {
		if sequence.(int32) < upTo {
			self.tree.Remove(sequence)
			self.capacity++
		}
	}

	self.ready.Broadcast()
}
