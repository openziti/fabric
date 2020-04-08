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
	"time"
)

type rxWindow struct {
	tree       *btree.Tree
	highWater  int32
	lastReport time.Time
	lastWater  int32
	lock       *sync.Mutex
}

func newRxWindow() *rxWindow {
	rxw := &rxWindow{
		tree:       btree.NewWith(10240, utils.Int32Comparator),
		highWater:  -1,
		lastReport: time.Now(),
		lastWater:  -1,
		lock:       new(sync.Mutex),
	}
	go rxw.monitor()
	return rxw
}

func (self *rxWindow) rx(m *message) []*message {
	self.lock.Lock()
	defer self.lock.Unlock()

	self.tree.Put(m.sequence, m)
	next := self.tree.LeftKey().(int32)
	var messages []*message
	for _, key := range self.tree.Keys() {
		if key.(int32) == next {
			m, _ := self.tree.Get(key)
			self.tree.Remove(key)
			messages = append(messages, m.(*message))
			self.highWater = key.(int32)
			next++

		} else {
			break
		}
	}

	self.report()

	return messages
}

func (self *rxWindow) monitor() {
	for {
		time.Sleep(1 * time.Second)

		self.lock.Lock()
		self.report()
		self.lock.Unlock()
	}
}

func (self *rxWindow) report() {
	if time.Since(self.lastReport).Milliseconds() >= 1000 && self.lastWater < self.highWater {
		// report(self.highWater)
		self.lastWater = self.highWater
		self.lastReport = time.Now()
	}
}
