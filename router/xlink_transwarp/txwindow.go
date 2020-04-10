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
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

type txWindow struct {
	tree       *btree.Tree
	highWater  int32
	capacity   int
	lastReport time.Time
	lock       *sync.Mutex
	available  *sync.Cond
	conn       *net.UDPConn
	peer       *net.UDPAddr
}

func newTxWindow(conn *net.UDPConn, peer *net.UDPAddr) *txWindow {
	txw := &txWindow{
		tree:      btree.NewWith(10240, utils.Int32Comparator),
		highWater: -1,
		capacity:  32,
		lock:      new(sync.Mutex),
		conn:      conn,
		peer:      peer,
	}
	txw.available = sync.NewCond(txw.lock)
	go txw.monitor()
	return txw
}

func (self *txWindow) tx(m *message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for self.capacity < 1 {
		self.available.Wait()
	}

	self.tree.Put(m.sequence, m)
	if m.sequence > self.highWater {
		self.highWater = m.sequence
	}

	logrus.Infof("[%d ^%d] ->", m.sequence, self.highWater)

	self.capacity--
}

func (self *txWindow) release(through int32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	oldCapacity := self.capacity
	for _, sequence := range self.tree.Keys() {
		if sequence.(int32) <= through {
			self.tree.Remove(sequence)
			self.capacity++
		}
	}
	self.lastReport = time.Now()

	if through < self.highWater {
		for i := through + 1; i <= self.highWater; i++ {
			if m, found := self.tree.Get(i); found {
				if err := writeMessage(m.(*message), nil, self.conn, self.peer); err == nil {
					logrus.Warnf("[*%d] ->", m.(*message).sequence)
				} else {
					logrus.Errorf("error retransmitting [%d] (%v)", m.(*message).sequence, err)
				}
			} else {
				logrus.Errorf("missing [%d] for retransmit", i)
			}
		}
	}

	logrus.Infof("[/%d ^%d (%d+>%d)] <=", through, self.highWater, oldCapacity, self.capacity)

	self.available.Broadcast()
}

func (self *txWindow) monitor() {
	for {
		time.Sleep(1 * time.Second)

		self.lock.Lock()
		if time.Since(self.lastReport).Seconds() >= 2 {
			logrus.Debugf("![/?]")
		}
		self.lock.Unlock()
	}
}
