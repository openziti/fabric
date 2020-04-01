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

func encodeWindowReport(sequence, lowWater, highWater, oops, count int32) (m *message, err error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, lowWater); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, highWater); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, oops); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, count); err != nil {
		return nil, err
	}

	m = &message{
		sequence:    sequence,
		fragment:    0,
		ofFragments: 1,
		messageType: WindowReport,
		payload:     payload.Bytes(),
	}

	return
}

func decodeWindowReport(m *message) (lowWater, highWater, oops, count int32, err error) {
	if len(m.payload) != 16 {
		return 0, 0, 0, 0, fmt.Errorf("expected 16 byte payload")
	}

	if value, err := readInt32(m.payload[0:4]); err == nil {
		lowWater = value
	} else {
		return 0, 0, 0, 0, err
	}
	if value, err := readInt32(m.payload[4:8]); err == nil {
		highWater = value
	} else {
		return 0, 0, 0, 0, err
	}
	if value, err := readInt32(m.payload[8:12]); err == nil {
		oops = value
	} else {
		return 0, 0, 0, 0, err
	}
	if value, err := readInt32(m.payload[12:16]); err == nil {
		count = value
	} else {
		return 0, 0, 0, 0, err
	}

	return
}

func encodeWindowSizeRequest(sequence, newSize int32) (*message, error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, newSize); err != nil {
		return nil, err
	}

	m := &message{
		sequence:    sequence,
		fragment:    0,
		ofFragments: 1,
		messageType: WindowSizeRequest,
		payload:     payload.Bytes(),
	}

	return m, nil
}

func decodeWindowSizeRequest(m *message) (int32, error) {
	newWindowSize, err := readInt32(m.payload[0:4])
	if err != nil {
		return 0, err
	}
	return newWindowSize, nil
}
