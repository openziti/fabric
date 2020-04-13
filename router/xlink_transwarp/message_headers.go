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

const (
	HeaderRtt = 0
)

/*
 * TRANSWARP v1 Headers Wire Format
 *
 * <key:uint8>                                  0
 * <length:uint8>                               1
 * <data>                                       2 -> (2 + length)
 */
func encodeHeaders(headers map[uint8][]byte) ([]byte, error) {
	data := new(bytes.Buffer)
	for k, v := range headers {
		if _, err := data.Write([]byte{k}); err != nil {
			return nil, err
		}
		if err := binary.Write(data, binary.LittleEndian, uint8(len(v))); err != nil {
			return nil, err
		}
		if n, err := data.Write(v); err == nil {
			if n != len(v) {
				return nil, fmt.Errorf("short header write")
			}
		} else {
			return nil, err
		}
	}
	return data.Bytes(), nil
}

func decodeHeaders(data []byte) (map[uint8][]byte, error) {
	headers := make(map[uint8][]byte)
	if len(data) > 0 && len(data) < 2 {
		return nil, fmt.Errorf("truncated header data")
	}
	i := 0
	for i < len(data) {
		key := data[i]
		length := data[i+1]
		if i+2+int(length) > len(data) {
			return nil, fmt.Errorf("short header data (%d > %d)", i+2+int(length), len(data))
		}
		headerData := data[i+2 : i+2+int(length)]
		headers[key] = headerData
		i += 2 + int(length)
	}
	return headers, nil
}
