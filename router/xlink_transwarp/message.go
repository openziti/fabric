package xlink_transwarp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/netfoundry/ziti-fabric/router/xgress"
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
	Ack
	Probe
	XgressPayload
	XgressAcknowledgement
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

func encodeAck(forSequence, windowSize int32) (m *message, err error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, forSequence); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, windowSize); err != nil {
		return nil, err
	}
	m = &message{
		sequence:    -1,
		fragment:    0,
		ofFragments: 1,
		messageType: Ack,
		payload:     payload.Bytes(),
	}
	return
}

func decodeAck(m *message) (forSequence, windowSize int32, err error) {
	if len(m.payload) < 8 {
		return -1, -1, fmt.Errorf("expected >=8 byte payload")
	}
	forSequence, err = readInt32(m.payload[0:4])
	if err != nil {
		return
	}
	windowSize, err = readInt32(m.payload[4:8])
	if err != nil {
		return
	}
	return
}

func encodeProbe() *message {
	return &message{
		sequence:    -1,
		fragment:    0,
		ofFragments: 1,
		messageType: Probe,
	}
}

/*
 * TRANSWARP v1 xgress.Payload Format
 *
 * <session_id_length:int32>					0  1  2  3
 * <flags:uint32>								4  5  6  7
 * <sequence:int32>								8  9 10 11
 * <data_length:int32>						   12 13 14 15
 * <session_id>								   16 -> (16 + session_id_length)
 * <data>                                      (16 + session_id_length) -> (16 + session_id_length + data_length)
 */
func encodeXgressPayload(p *xgress.Payload, sequence int32) (m *message, err error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, int32(len(p.SessionId))); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, p.Flags); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, p.Sequence); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, int32(len(p.Data))); err != nil {
		return nil, err
	}
	if _, err := payload.Write([]byte(p.SessionId)); err != nil {
		return nil, err
	}
	if _, err := payload.Write(p.Data); err != nil {
		return nil, err
	}

	m = &message{
		sequence:    sequence,
		fragment:    0,
		ofFragments: 1,
		messageType: XgressPayload,
		headers:     p.Headers,
		payload:     payload.Bytes(),
	}

	return
}

func decodeXgressPayload(m *message) (p *xgress.Payload, err error) {
	sessionIdLength, err := readInt32(m.payload[0:4])
	if err != nil {
		return nil, err
	}
	flags, err := readUint32(m.payload[4:8])
	if err != nil {
		return nil, err
	}
	sequence, err := readInt32(m.payload[8:12])
	if err != nil {
		return nil, err
	}
	dataLength, err := readInt32(m.payload[12:16])
	if err != nil {
		return nil, err
	}
	sessionId := m.payload[16 : 16+sessionIdLength]
	data := m.payload[16+sessionIdLength : 16+sessionIdLength+dataLength]

	p = &xgress.Payload{
		Header: xgress.Header{
			SessionId: string(sessionId),
			Flags:     flags,
		},
		Sequence: sequence,
		Headers:  m.headers,
		Data:     data,
	}

	return p, nil
}

/*
 * TRANSWARP v1 xgress.Acknowledgement Format
 *
 * <session_id_length:int32>					0  1  2  3
 * <flags:uint32>								4  5  6  7
 * <sequence_ids_count:int32>					8  9 10 11
 * <session_id>									12 -> (12 + session_id_length)
 * <sequence_ids:[]int32>						(12 + session_id_length) -> ((12 + session_id_length) + (4 * sequence_ids_count))
 */
func encodeXgressAcnowledgement(a *xgress.Acknowledgement, sequence int32) (m *message, err error) {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, int32(len(a.SessionId))); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, a.Flags); err != nil {
		return nil, err
	}
	if err := binary.Write(payload, binary.LittleEndian, int32(len(a.Sequence))); err != nil {
		return nil, err
	}
	if _, err := payload.Write([]byte(a.SessionId)); err != nil {
		return nil, err
	}
	for _, sequenceId := range a.Sequence {
		if err := binary.Write(payload, binary.LittleEndian, sequenceId); err != nil {
			return nil, err
		}
	}

	m = &message{
		sequence:    sequence,
		fragment:    0,
		ofFragments: 1,
		messageType: XgressAcknowledgement,
		payload:     payload.Bytes(),
	}

	return
}

func decodeXgressAcknowledgement(m *message) (a *xgress.Acknowledgement, err error) {
	sessionIdLength, err := readInt32(m.payload[0:4])
	if err != nil {
		return nil, err
	}
	flags, err := readUint32(m.payload[4:8])
	if err != nil {
		return nil, err
	}
	sequenceIdsCount, err := readInt32(m.payload[8:12])
	if err != nil {
		return nil, err
	}
	sessionId := m.payload[12 : 12+sessionIdLength]
	nextSequenceId := 12 + sessionIdLength
	var sequenceIds []int32
	for i := 0; i < int(sequenceIdsCount); i++ {
		sequenceId, err := readInt32(m.payload[nextSequenceId : nextSequenceId+4])
		if err != nil {
			return nil, err
		}
		sequenceIds = append(sequenceIds, sequenceId)
		nextSequenceId += 4
	}

	a = &xgress.Acknowledgement{
		Header: xgress.Header{
			SessionId: string(sessionId),
			Flags:     flags,
		},
		Sequence: sequenceIds,
	}

	return
}
