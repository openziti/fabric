package xlink_transwarp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

type message struct {
	sequence    int32
	fragment    uint8
	ofFragments uint8
	messageType messageType
	headers     map[uint8][]byte
	payload     []byte
}

/**
 * TRANSWARP v1 Wire Format
 *
 * // --- message section --------------------------------------------------------------------------------- //
 *
 * <version:[]byte>								0  1  2  3
 * <sequence:int32> 							4  5  6  7
 * <fragment:uint8>								8
 * <of_fragments:uint8>							9
 * <content_type:uint8>							10
 * <headers_length:uint16>						11 12
 * <payload_length:uint16> 						13 14
 *
 * // --- data section ------------------------------------------------------------------------------------ //
 *
 * <headers>									15 -> (15 + headers_length)
 * <payload>									(15 + headers_length) -> (15 + headers_length + payload_length)
 */
var magicV1 = []byte{0x01, 0x02, 0x02, 0x00}

const messageSectionLength = 15

type messageType uint8

const (
	Hello messageType = iota
	Ping
	Payload
	Acknowledgement
	WindowReport
	WindowRequest
)

const timeoutSeconds = 5
const mss = 1472
const noReplyFor = -1

func encodeMessage(m *message) ([]byte, error) {
	if m == nil {
		return nil, fmt.Errorf("nil message")
	}

	data := new(bytes.Buffer)

	data.Write(magicV1)
	if err := binary.Write(data, binary.LittleEndian, m.sequence); err != nil {
		return nil, fmt.Errorf("sequence write (%w)", err)
	}
	if _, err := data.Write([]byte{m.fragment, m.ofFragments, uint8(m.messageType)}); err != nil {
		return nil, fmt.Errorf("fragments/type write (%w)", err)
	}
	var headers []byte
	if m.headers != nil {
		var err error
		headers, err = encodeHeaders(m.headers)
		if err != nil {
			return nil, fmt.Errorf("encoding headers (%w)", err)
		}
	}
	headersLength := len(headers)
	if err := binary.Write(data, binary.LittleEndian, uint16(headersLength)); err != nil { // headers length
		return nil, fmt.Errorf("headers length write (%w)", err)
	}
	payloadLength := len(m.payload)
	if err := binary.Write(data, binary.LittleEndian, uint16(payloadLength)); err != nil {
		return nil, fmt.Errorf("payload length write (%w)", err)
	}
	if headersLength > 0 {
		n, err := data.Write(headers)
		if err != nil {
			return nil, fmt.Errorf("headers write (%w)", err)
		}
		if n != headersLength {
			return nil, fmt.Errorf("short headers write [%d != %d]", n, headersLength)
		}
	}
	if payloadLength > 0 {
		n, err := data.Write(m.payload)
		if err != nil {
			return nil, fmt.Errorf("payload write (%w)", err)
		}
		if n != payloadLength {
			return nil, fmt.Errorf("short payload write [%d != %d]", n, payloadLength)
		}
	}

	buffer := make([]byte, data.Len())
	_, err := data.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("error reading buffer (%w)", err)
	}
	/*
		if n > mss {
			return nil, fmt.Errorf("message too long [%d]", n)
		}
	*/

	return buffer, nil
}

func decodeMessage(data []byte) (*message, error) {
	m := &message{}
	if len(data) < messageSectionLength {
		return nil, fmt.Errorf("short read")
	}
	for i := 0; i < len(magicV1); i++ {
		if data[i] != magicV1[i] {
			return nil, fmt.Errorf("bad magic")
		}
	}
	sequence, err := readInt32(data[4:8])
	if err != nil {
		return nil, fmt.Errorf("error reading sequence (%w)", err)
	}
	m.sequence = sequence

	m.fragment = data[8]
	m.ofFragments = data[9]
	m.messageType = messageType(data[10])

	headersLength, err := readUint16(data[11:13])
	if err != nil {
		return nil, fmt.Errorf("error reading headers length (%w)", err)
	}
	if headers, err := decodeHeaders(data[15 : 15+headersLength]); err == nil {
		m.headers = headers
	} else {
		return nil, fmt.Errorf("headers error (%w)", err)
	}

	payloadLength, err := readUint16(data[13:15])
	if err != nil {
		return nil, fmt.Errorf("error reading payload length (%w)", err)
	}
	m.payload = data[15+headersLength : 15+headersLength+payloadLength]

	return m, nil
}

func readInt32(data []byte) (ret int32, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func writeInt32(value int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readInt64(data []byte) (ret int64, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func writeInt64(value int64) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readUint32(data []byte) (ret uint32, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.LittleEndian, &ret)
	return
}

func readUint16(data []byte) (ret uint16, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.LittleEndian, &ret)
	return
}

const millisInSecond = 1000
const nsInSecond = 1000000

func fromMilliseconds(ms int64) time.Time {
	return time.Unix(ms/int64(millisInSecond), (ms%int64(millisInSecond))*int64(nsInSecond))
}
