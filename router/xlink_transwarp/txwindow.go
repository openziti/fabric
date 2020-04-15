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

const startingWindowCapacity = 32

type txWindow struct {
	tree        *btree.Tree
	lock        *sync.Mutex
	ready       *sync.Cond
	capacity    int
	monitorCtrl chan txMonitorRequest
	conn        *net.UDPConn
	peer        *net.UDPAddr
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

func newTxWindow(conn *net.UDPConn, peer *net.UDPAddr) *txWindow {
	txw := &txWindow{
		tree:        btree.NewWith(10240, utils.Int32Comparator),
		lock:        new(sync.Mutex),
		capacity:    startingWindowCapacity,
		monitorCtrl: make(chan txMonitorRequest, 1),
		conn:        conn,
		peer:        peer,
	}
	txw.ready = sync.NewCond(txw.lock)
	go txw.monitor()
	return txw
}

func (self *txWindow) tx(m *message) {
	self.lock.Lock()
	defer self.lock.Unlock()

	for self.capacity < 1 {
		self.ready.Wait()
	}

	self.tree.Put(m.sequence, m)
	self.monitorCtrl <- txMonitorRequest{monitor: true, m: m}
	self.capacity--
}

func (self *txWindow) ack(forSequence int32) {
	self.lock.Lock()
	defer self.lock.Unlock()

	if m, found := self.tree.Get(forSequence); found {
		self.monitorCtrl <- txMonitorRequest{monitor: false, m: m.(*message)}
		self.tree.Remove(forSequence)
		self.capacity++
		self.ready.Broadcast()

	} else {
		logrus.Warnf("repeated ack forSequence [%d]", forSequence)
	}
}

func (self *txWindow) monitor() {
	logrus.Infof("started")
	defer logrus.Warnf("exited")

	var window []txMonitorState
	for {
		if len(window) > 0 {
			logrus.Debugf("waiting for [%s]", time.Until(window[0].timeout))

			select {
			case request, ok := <-self.monitorCtrl:
				if ok {
					window = updateMonitorList(request, window)
				} else {
					return
				}

			case <-time.After(time.Until(window[0].timeout)):
				if err := writeMessage(window[0].m, nil, self.conn, self.peer); err == nil {
					logrus.Warnf("[* %d] =>", window[0].m.sequence)
				} else {
					logrus.Errorf("error retransmitting [#%d] (%v)", window[0].m.sequence, err)
				}

				window[0].retries++
				window[0].timeout = time.Now().Add(time.Duration(50 * window[0].retries) * time.Millisecond)
				sort.Slice(window, func(i, j int) bool {
					return window[i].timeout.Before(window[j].timeout)
				})
			}
		} else {
			select {
			case request, ok := <-self.monitorCtrl:
				if ok {
					window = updateMonitorList(request, window)
				} else {
					return
				}
			}
		}
	}
}

func updateMonitorList(request txMonitorRequest, window []txMonitorState) []txMonitorState {
	if request.monitor {
		// Add node to monitor list
		window = append(window, txMonitorState{timeout: time.Now().Add(200 * time.Millisecond), m: request.m})
		sort.Slice(window, func(i, j int) bool {
			return window[i].timeout.Before(window[j].timeout)
		})
	} else {
		// Remove node from monitor list
		i := -1
		for j, node := range window {
			if node.m == request.m {
				i = j
				break
			}
		}
		if i > -1 {
			window = append(window[:i], window[i+1:]...)
		}
	}
	return window
}
