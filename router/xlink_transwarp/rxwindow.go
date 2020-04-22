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
	"strconv"
	"sync"
)

type rxWindow struct {
	tree      *btree.Tree
	lock      *sync.Mutex
	highWater int32
	xli       *impl
	trace     chan interface{}
}

func newRxWindow(xli *impl, trace chan interface{}) *rxWindow {
	rxw := &rxWindow{
		tree:      btree.NewWith(10240, utils.Int32Comparator),
		lock:      new(sync.Mutex),
		highWater: -1,
		xli:       xli,
		trace:     trace,
	}
	return rxw
}

func (self *rxWindow) rx(m *message) []*message {
	self.lock.Lock()
	defer self.lock.Unlock()

	if m.sequence > self.highWater {
		self.tree.Put(m.sequence, m)
	}
	self.ack(m)

	var output []*message
	if self.tree.Size() > 0 {
		next := self.tree.LeftKey().(int32)
		if next <= self.highWater {
			logrus.Fatalf("how?")
		}
		for _, key := range self.tree.Keys() {
			if key.(int32) == next {
				m, _ := self.tree.Get(key)
				self.tree.Remove(key)
				output = append(output, m.(*message))
				self.highWater = key.(int32)
				next++
			} else {
				break
			}
		}
	}

	if len(output) != 1 || (len(output) > 0 && output[0].sequence != m.sequence) {
		msg := fmt.Sprintf("#%d, ^%d, tree %v, outputting [", m.sequence, self.highWater, self.tree.Keys())
		for _, m := range output {
			msg += " " + strconv.Itoa(int(m.sequence))
		}
		msg += " ]"
		logrus.Info(msg)
	}

	self.trace <- buildTraceRx(self)

	return output
}

func (self *rxWindow) ack(m *message) {
	if err := writeAck(m.sequence, startingWindowCapacity-int32(self.tree.Size()), self.xli); err == nil {
		logrus.Debugf("[@ %d] ->", m.sequence)
	} else {
		logrus.Errorf("error sending ack for [#%d] (%v)", m.sequence, err)
	}
}
