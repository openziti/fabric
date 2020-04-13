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
	"bytes"
	"encoding/binary"
	"fmt"
)

func encodeWindowReport(highWater int32, rtt []byte) (m *message, err error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, highWater); err != nil {
		return nil, err
	}
	if _, err := payload.Write(rtt); err != nil {
		return nil, err
	}
	m = &message{
		sequence:    -1,
		fragment:    0,
		ofFragments: 1,
		messageType: WindowReport,
		payload:     payload.Bytes(),
	}
	return
}

func decodeWindowReport(m *message) (highWater int32, rtt []byte, err error) {
	if len(m.payload) < 4 {
		return -1, nil, fmt.Errorf("expected >=4 byte payload")
	}
	if value, err := readInt32(m.payload[0:4]); err == nil {
		highWater = value
	} else {
		return -1, nil, err
	}
	rtt = m.payload[4:]
	return
}

func encodeWindowRequest() *message {
	return &message{
		sequence:    -1,
		fragment:    0,
		ofFragments: 1,
		messageType: WindowRequest,
	}
}