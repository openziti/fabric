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
	"github.com/netfoundry/ziti-foundation/identity/identity"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"time"
)

type traceRxWindow struct {
	tree      []int32
	highWater int32
}

type traceTxWindow struct {
	tree     []int32
	capacity int
}

type traceTx struct {
	sequence   int32
	retransmit bool
}

type traceTxAck struct {
	forSequence int32
}

type traceRx struct {
	sequence int32
	discard  bool
}

type traceRxAck struct {
	forSequence int32
}

func buildTraceRx(rxw *rxWindow) traceRxWindow {
	var trx traceRxWindow
	for _, sequence := range rxw.tree.Keys() {
		trx.tree = append(trx.tree, sequence.(int32))
	}
	trx.highWater = rxw.highWater
	return trx
}

func buildTraceTx(txw *txWindow) traceTxWindow {
	var ttx traceTxWindow
	for _, sequence := range txw.tree.Keys() {
		ttx.tree = append(ttx.tree, sequence.(int32))
	}
	ttx.capacity = txw.capacity
	return ttx
}

func buildTraceTxMsg(m *message, retransmit bool) traceTx {
	return traceTx{sequence: m.sequence, retransmit: retransmit}
}

type traceController struct {
	id *identity.TokenId
	in chan interface{}
	oF *os.File
}

func newTrace(id *identity.TokenId, in chan interface{}) (*traceController, error) {
	oF, err := ioutil.TempFile(".", fmt.Sprintf("%s-*.twtrace", id.Token))
	if err != nil {
		return nil, err
	}
	return &traceController{id: id, in: in, oF: oF}, nil
}

func (self *traceController) run() {
	logrus.Infof("started")
	defer logrus.Errorf("exited")

	for {
		var out string
		select {
		case msg := <-self.in:
			switch t := msg.(type) {
			case traceRxWindow:
				out = fmt.Sprintf("%-20s [", "rxWindow")
				for i, sequence := range t.tree {
					if i > 0 {
						out += ", "
					}
					out += "#" + strconv.Itoa(int(sequence))
				}
				out += fmt.Sprintf("], ^#%d", t.highWater)

			case traceTxWindow:
				out = fmt.Sprintf("%-20s [", "txWindow")
				for i, sequence := range t.tree {
					if i > 0 {
						out += ", "
					}
					out += "#" + strconv.Itoa(int(sequence))
				}
				out += fmt.Sprintf("], ~(%d)", t.capacity)

			case traceTx:
				out = fmt.Sprintf("%-20s #%d", "tx", t.sequence)
				if t.retransmit {
					out += " (retransmit)"
				}

			case traceTxAck:
				out = fmt.Sprintf("%-20s #%d", "txAck", t.forSequence)

			case traceRx:
				out = fmt.Sprintf("%-20s #%d", "rx", t.sequence)
				if t.discard {
					out += "(discard)"
				}

			case traceRxAck:
				out = fmt.Sprintf("%-20s #%d", "rxAck", t.forSequence)

			default:
				out = fmt.Sprintf("no trace decoder [%s]", reflect.TypeOf(msg))
			}
		}
		fmt.Fprintf(self.oF, "%d: %s\n", time.Now().UnixNano()/int64(time.Millisecond), out)
	}
}
