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
	"sort"
	"sync"
	"time"
)

const startingWindowCapacity = 6
const retransmissionDelay = 20

type txWindow struct {
	tree             *btree.Tree
	lock             *sync.Mutex
	capacity         int
	capacityReady    *sync.Cond
	monitorQueue     []*txMonitorState
	monitorSubject   *txMonitorState
	monitorCancelled bool
	monitorReady     *sync.Cond
	conn             *net.UDPConn
	peer             *net.UDPAddr
	trace            chan interface{}
}

type txMonitorRequest struct {
	monitor bool
	m       *message
}

type txMonitorState struct {
	timeout time.Time
	retries int
	m       *message
}

func newTxWindow(conn *net.UDPConn, peer *net.UDPAddr, trace chan interface{}) *txWindow {
	txw := &txWindow{
		tree:     btree.NewWith(10240, utils.Int32Comparator),
		lock:     new(sync.Mutex),
		capacity: startingWindowCapacity,
		conn:     conn,
		peer:     peer,
		trace:    trace,
	}
	txw.capacityReady = sync.NewCond(txw.lock)
	txw.monitorReady = sync.NewCond(txw.lock)
	go txw.retransmitter()
	return txw
}

func (self *txWindow) tx(m *message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for self.capacity < 1 {
		self.capacityReady.Wait()
	}

	self.tree.Put(m.sequence, m)
	self.addMonitored(m)
	self.capacity--

	self.trace <- buildTraceTx(self)
}

func (self *txWindow) ack(forSequence int32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if m, found := self.tree.Get(forSequence); found {
		self.removeMonitored(m.(*message))
		self.tree.Remove(forSequence)
		self.capacity++
		self.capacityReady.Broadcast()

	} else {
		logrus.Warnf("[!@ %d] <=", forSequence)
	}

	self.trace <- buildTraceTx(self)
}

func (self *txWindow) addMonitored(m *message) {
	self.monitorQueue = append(self.monitorQueue, &txMonitorState{timeout: time.Now().Add(retransmissionDelay * time.Millisecond), m: m})
	sort.Slice(self.monitorQueue, func(i, j int) bool {
		return self.monitorQueue[i].timeout.Before(self.monitorQueue[j].timeout)
	})
	self.monitorReady.Signal()
}

func (self *txWindow) removeMonitored(m *message) {
	i := -1
	for j, monitor := range self.monitorQueue {
		if monitor.m == m {
			i = j
			break
		}
	}
	if i > -1 {
		self.monitorQueue = append(self.monitorQueue[:i], self.monitorQueue[i+1:]...)
	}
	if self.monitorSubject != nil && self.monitorSubject.m == m {
		self.monitorCancelled = true
	}
}

func (self *txWindow) retransmitter() {
	for {
		var timeout time.Duration
		{
			self.lock.Lock()

			for len(self.monitorQueue) < 1 {
				self.monitorReady.Wait()
			}
			self.monitorSubject = self.monitorQueue[0]
			timeout = time.Until(self.monitorSubject.timeout)
			self.monitorCancelled = false

			self.lock.Unlock()
		}

		time.Sleep(timeout)

		{
			self.lock.Lock()

			if !self.monitorCancelled {
				if err := writeMessage(self.monitorSubject.m, nil, self.conn, self.peer); err == nil {
					logrus.Warnf("[* %d] =>", self.monitorSubject.m.sequence)
				} else {
					logrus.Errorf("[!* %d] => (%v)", self.monitorSubject.m.sequence, err)
				}

				self.monitorSubject.timeout = time.Now().Add(retransmissionDelay * time.Millisecond)
				sort.Slice(self.monitorQueue, func(i, j int) bool {
					return self.monitorQueue[i].timeout.Before(self.monitorQueue[j].timeout)
				})
			} else {
				logrus.Debugf("[X* %d] =>", self.monitorSubject.m.sequence)
			}

			self.lock.Unlock()
		}
	}
}
