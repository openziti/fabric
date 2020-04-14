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
	"github.com/netfoundry/ziti-fabric/router/xgress"
	"github.com/netfoundry/ziti-fabric/router/xlink"
	"github.com/netfoundry/ziti-foundation/identity/identity"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

/*
 * xlink.Xlink
 */
func (self *impl) Id() *identity.TokenId {
	return self.id
}

func (self *impl) SendPayload(payload *xgress.Payload) error {
	return writeXgressPayload(payload, self)
}

func (self *impl) SendAcknowledgement(acknowledgement *xgress.Acknowledgement) error {
	return writeXgressAcknowledgement(acknowledgement, self)
}

func (self *impl) Close() error {
	return self.conn.Close()
}

/*
 * xlink_transwarp.MessageHandler
 */
func (self *impl) HandlePing(sequence int32, replyFor int32) {
	if replyFor == -1 {
		if err := self.sendPingReply(sequence); err != nil {
			logrus.Errorf("error sending ping (%v)", err)
		}
	} else {
		self.receivePing(replyFor)
	}
}

func (self *impl) HandlePayload(p *xgress.Payload) {
	if err := self.forwarder.ForwardPayload(xgress.Address(self.id.Token), p); err != nil {
		logrus.Errorf("[l/%s] => error forwarding payload (%v)", self.id.Token, err)
	}
}

func (self *impl) HandleAcknowledgement(a *xgress.Acknowledgement) {
	if err := self.forwarder.ForwardAcknowledgement(xgress.Address(self.id.Token), a); err != nil {
		logrus.Errorf("[l/%s] => error forwarding acknowledgement (%v)", self.id.Token, err)
	}
}

func (self *impl) HandleWindowReport(forSequence, windowSize int32) {
	logrus.Infof("{%s} [#%d, ~%d] <= ", self.id.Token, forSequence, windowSize)
}

func (self *impl) HandleWindowRequest() {
	logrus.Warnf("not implemented")
}

/*
 * impl
 */
func (self *impl) listener() {
	for {
		if m, peer, err := readMessage(self.conn); err == nil {
			if m.messageType != Ack && m.messageType != Probe {
				mrs := self.rxWindow.rx(m)
				for _, mr := range mrs {
					if err := handleMessage(mr, self.conn, peer, self); err != nil {
						logrus.Errorf("error handling message from [%s] (%v)", peer, err)
					}
				}
 			} else {
 				if err := handleMessage(m, self.conn, peer, self); err != nil {
					logrus.Errorf("error handling message from [%s] (%v)", peer, err)
				}
			}
		}
	}
}

func (self *impl) pinger() {
	for {
		time.Sleep(pingCycleDelayMs * time.Millisecond)
		if time.Since(self.lastPingTx).Milliseconds() >= pingDelayMs {
			if err := self.sendPingRequest(); err != nil {
				logrus.Errorf("error sending ping request (%v)", err)
			}
		}
		logrus.Debugf("time since last ping receipt [%d ms.]", time.Since(self.lastPingRx).Milliseconds())
	}
}

func (self *impl) sendPingRequest() error {
	if sequence, err := writePing(noReplyFor, self); err == nil {
		self.lastPingTxSequence = sequence
		self.lastPingTx = time.Now()

		logrus.Infof("[ping:%d] => %s", sequence, self.peer)

		return nil

	} else {
		return fmt.Errorf("error sending ping (%w)", err)
	}
}

func (self *impl) sendPingReply(forSequence int32) error {
	if _, err := writePing(forSequence, self); err == nil {
		logrus.Infof("[ping:%d] <= %s", forSequence, self.peer)
		return nil
	} else {
		return fmt.Errorf("error sending ping reply to [%s] (%w)", self.peer, err)
	}
}

func (self *impl) receivePing(replyFor int32) {
	if replyFor == self.lastPingTxSequence {
		self.lastPingRx = time.Now()
		logrus.Debugf("received outstanding ping for [l/%s]", self.id.Token)
	}
}

func (self *impl) nextSequence() int32 {
	self.sequenceLock.Lock()
	defer self.sequenceLock.Unlock()

	sequence := self.sequence
	self.sequence++
	return sequence
}

func newImpl(id *identity.TokenId, conn *net.UDPConn, peer *net.UDPAddr, f xlink.Forwarder) *impl {
	now := time.Now()
	xlinkImpl := &impl{
		id:         id,
		conn:       conn,
		peer:       peer,
		lastPingRx: now,
		lastPingTx: now,
		forwarder:  f,
		rxWindow:   newRxWindow(conn, peer),
		txWindow:   newTxWindow(conn, peer),
	}
	return xlinkImpl
}

type impl struct {
	id                 *identity.TokenId
	conn               *net.UDPConn
	peer               *net.UDPAddr
	sequence           int32
	sequenceLock       sync.Mutex
	lastPingRx         time.Time
	lastPingTx         time.Time
	lastPingTxSequence int32
	forwarder          xlink.Forwarder
	rxWindow           *rxWindow
	txWindow           *txWindow
}

const pingDelayMs = 5000
const pingCycleDelayMs = 500
