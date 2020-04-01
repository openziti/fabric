/*
	Copyright NetFoundry, Inc.

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
	"github.com/sirupsen/logrus"
	"math"
	"sync"
	"time"
)

func (self *transmitBuffer) ack(gaps int) {
	self.cond.L.Lock()
	if gaps > 0 {
		if self.windowSize > 10 {
			self.windowSize /= 2
		}
	}
	self.windowContains = 0
	self.cond.L.Unlock()
}

func (self *transmitBuffer) accept(m *message) {
	self.cond.L.Lock()
	for self.windowContains >= self.windowSize {
		self.cond.Wait()
	}
	self.highWater = m.sequence
	self.windowContains++
	self.cond.L.Unlock()
}

func (self *transmitBuffer) release() {
	self.cond.L.Lock()
	self.windowContains = 0
	self.cond.L.Unlock()
	self.cond.Broadcast()
	logrus.Warnf("released")
}

func newTransmitBuffer() *transmitBuffer {
	txb := &transmitBuffer{
		windowSize: startingWindowSize,
		highWater:  -1,
	}
	txb.lock = new(sync.Mutex)
	txb.cond = sync.NewCond(txb.lock)
	return txb
}

type transmitBuffer struct {
	windowContains int
	windowSize     int
	highWater      int32
	lock           *sync.Mutex
	cond           *sync.Cond
}

func (self *receiveBuffer) receive(m *message) {
	self.lock.Lock()
	if m.sequence != self.highWater+1 {
		self.oops += int32(math.Abs(float64(m.sequence - (self.highWater + 1))))
	}
	if m.sequence < self.lowWater {
		self.lowWater = m.sequence
	}
	if m.sequence > self.highWater {
		self.highWater = m.sequence
	}
	self.count++
	if self.count >= self.windowSize {
		if err := writeWindowReport(self.xlinkImpl.nextSequence(), self.lowWater, self.highWater, self.oops, self.count, self.xlinkImpl.txBuffer, self.xlinkImpl.conn, self.xlinkImpl.peer); err == nil {
			if self.oops > 0 {
				logrus.Infof("[lw:%d, hw:%d, oo:%d, c:%d] => [%s]", self.lowWater, self.highWater, self.oops, self.count, self.xlinkImpl.peer)
			}
		} else {
			logrus.Errorf("error writing window report (%v)", err)
		}
		self.lowWater = self.highWater
		self.count = 0
		self.oops = 0
		self.lastReport = time.Now()
	}
	self.lock.Unlock()
}

func (self *receiveBuffer) acker() {
	for {
		time.Sleep(3 * time.Second)

		self.lock.Lock()
		if time.Since(self.lastReport).Milliseconds() > 1000 {
			logrus.Infof("sending transwarp window report")

			if err := writeWindowReport(self.xlinkImpl.nextSequence(), self.lowWater, self.highWater, self.oops, self.count, self.xlinkImpl.txBuffer, self.xlinkImpl.conn, self.xlinkImpl.peer); err == nil {
				if self.oops > 0 {
					logrus.Infof("[lw:%d, hw:%d, oo:%d, c:%d] => [%s]", self.lowWater, self.highWater, self.oops, self.count, self.xlinkImpl.peer)
				}
			} else {
				logrus.Errorf("error writing window report (%v)", err)
			}

			self.lastReport = time.Now()
		}
		self.lock.Unlock()
	}
}

func newReceiveBuffer(xlinkImpl *impl) *receiveBuffer {
	rxb := &receiveBuffer{
		lastReport: time.Now(),
		windowSize: startingWindowSize,
		xlinkImpl:  xlinkImpl,
	}
	rxb.lock = new(sync.Mutex)
	//go rxb.acker()
	return rxb
}

type receiveBuffer struct {
	lowWater   int32
	highWater  int32
	count      int32
	oops       int32
	windowSize int32
	lastReport time.Time
	xlinkImpl  *impl
	lock       *sync.Mutex
}

const startingWindowSize = 10
