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
	"fmt"
	"github.com/emirpasic/gods/trees/btree"
	"github.com/emirpasic/gods/utils"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type rxWindow struct {
	tree       *btree.Tree
	highWater  int32
	lastReport time.Time
	lastWater  int32
	lock       *sync.Mutex
	conn       *net.UDPConn
	peer       *net.UDPAddr
}

func newRxWindow(conn *net.UDPConn, peer *net.UDPAddr) *rxWindow {
	rxw := &rxWindow{
		tree:       btree.NewWith(10240, utils.Int32Comparator),
		highWater:  -1,
		lastReport: time.Now(),
		lastWater:  -1,
		lock:       new(sync.Mutex),
		conn:       conn,
		peer:       peer,
	}
	go rxw.monitor()
	return rxw
}

func (self *rxWindow) rx(m *message) []*message {
	self.lock.Lock()
	defer self.lock.Unlock()

	if m.sequence > self.highWater {
		self.tree.Put(m.sequence, m)
	} else {
		logrus.Warnf("[v%d] <-", m.sequence)
	}

	var messages []*message
	if self.tree.Size() > 0 {
		next := self.highWater+1
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
	}

	count := len(messages)
	if count != 1 {
		out := "<- ["
		for _, m := range messages {
			out += fmt.Sprintf(" %d", m.sequence)
		}
		out += fmt.Sprintf(" ] <- [%d]", m.sequence)
		if count == 0 {
			logrus.Errorf(out)
		} else {
			logrus.Warnf(out)
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
	if time.Since(self.lastReport).Milliseconds() >= 1000 || self.highWater-self.lastWater >= 6 {
		if err := writeWindowReport(self.highWater, self.conn, self.peer); err == nil {
			self.lastWater = self.highWater
			self.lastReport = time.Now()
			logrus.Infof("[/%d] =>", self.highWater)
		}
	}
}
