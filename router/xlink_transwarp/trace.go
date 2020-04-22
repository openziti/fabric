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
	"reflect"
	"strconv"
)

type traceRx struct {
	tree      []int32
	highWater int32
}

type traceTx struct {
	tree     []int32
	capacity int
}

func buildTraceRx(rxw *rxWindow) traceRx {
	var trx traceRx
	for _, sequence := range rxw.tree.Keys() {
		trx.tree = append(trx.tree, sequence.(int32))
	}
	trx.highWater = rxw.highWater
	return trx
}

func buildTraceTx(txw *txWindow) traceTx {
	var ttx traceTx
	for _, sequence := range txw.tree.Keys() {
		ttx.tree = append(ttx.tree, sequence.(int32))
	}
	ttx.capacity = txw.capacity
	return ttx
}

type traceController struct {
	id *identity.TokenId
	in chan interface{}
}

func newTrace(id *identity.TokenId, in chan interface{}) *traceController {
	return &traceController{id: id, in: in}
}

func (self *traceController) run() {
	logrus.Infof("started")
	defer logrus.Errorf("exited")

	for {
		var out string
		select {
		case msg := <-self.in:
			switch t := msg.(type) {
			case traceRx:
				out = "traceRx{tree["
				for i, sequence := range t.tree {
					if i > 0 {
						out += ", "
					}
					out += strconv.Itoa(int(sequence))
				}
				out += fmt.Sprintf("], highWater[%d]", t.highWater)
				out += "}"

			case traceTx:
				out = "traceTx{tree["
				for i, sequence := range t.tree {
					if i > 0 {
						out += ", "
					}
					out += strconv.Itoa(int(sequence))
				}
				out += fmt.Sprintf("], capacity[%d]", t.capacity)
				out += "}"

			default:
				out = fmt.Sprintf("no trace decoder [%s]", reflect.TypeOf(msg))
			}
		}
		logrus.Infof("{%s}: %s", self.id.Token, out)
	}
}
