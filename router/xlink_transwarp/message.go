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
	"github.com/netfoundry/ziti-fabric/router/xgress"
	"github.com/netfoundry/ziti-foundation/identity/identity"
	"github.com/netfoundry/ziti-foundation/transport/udp"
	"net"
	"time"
)

type HelloHandler interface {
	HandleHello(linkId *identity.TokenId, conn *net.UDPConn, addr *net.UDPAddr)
}

type MessageHandler interface {
	HandlePing(sequence int32, replyFor int32, conn *net.UDPConn, addr *net.UDPAddr)
	HandlePayload(p *xgress.Payload, sequence int32, conn *net.UDPConn, addr *net.UDPAddr)
	HandleAcknowledgement(a *xgress.Acknowledgement, sequence int32, conn *net.UDPConn, addr *net.UDPAddr)
	HandleWindowReport(lowWater, highWater, gaps, count int32, conn *net.UDPConn, addr *net.UDPAddr)
	HandleWindowSizeRequest(newWindowSize int32, conn *net.UDPConn, addr *net.UDPAddr)
}

func writeHello(linkId *identity.TokenId, conn *net.UDPConn, peer *net.UDPAddr) error {
	payload := new(bytes.Buffer)
	payload.Write([]byte(linkId.Token))
	return writeMessage(&message{
		sequence:    -1,
		fragment:    0,
		ofFragments: 1,
		messageType: Hello,
		payload:     payload.Bytes(),
	}, conn, peer)
}

func writePing(sequence, replyFor int32, conn *net.UDPConn, peer *net.UDPAddr) error {
	payload := new(bytes.Buffer)
	if err := binary.Write(payload, binary.LittleEndian, replyFor); err != nil {
		return fmt.Errorf("reply for write (%w)", err)
	}
	return writeMessage(&message{
		sequence:    sequence,
		fragment:    0,
		ofFragments: 1,
		messageType: Ping,
		payload:     payload.Bytes(),
	}, conn, peer)
}

func writePayload(sequence int32, p *xgress.Payload, conn *net.UDPConn, peer *net.UDPAddr) error {
	m, err := encodePayload(p, sequence)
	if err != nil {
		return err
	}
	return writeMessage(m, conn, peer)
}

func writeAcknowledgement(sequence int32, a *xgress.Acknowledgement, conn *net.UDPConn, peer *net.UDPAddr) error {
	m, err := encodeAcnowledgement(a, sequence)
	if err != nil {
		return err
	}
	return writeMessage(m, conn, peer)
}

func writeWindowReport(sequence int32, lowWater, highWater, gaps, count int32, conn *net.UDPConn, peer *net.UDPAddr) error {
	m, err := encodeWindowReport(sequence, lowWater, highWater, gaps, count)
	if err != nil {
		return err
	}
	return writeMessage(m, conn, peer)
}

func writeWindowSizeRequest(sequence, newWindowSize int32, conn *net.UDPConn, peer *net.UDPAddr) error {
	m, err := encodeWindowSizeRequest(sequence, newWindowSize)
	if err != nil {
		return err
	}
	return writeMessage(m, conn, peer)
}

func writeMessage(m *message, conn *net.UDPConn, peer *net.UDPAddr) error {
	data, err := encodeMessage(m)
	if err != nil {
		return fmt.Errorf("error encoding bitstream (%w)", err)
	}

	if err := conn.SetWriteDeadline(time.Now().Add(timeoutSeconds * time.Second)); err != nil {
		return fmt.Errorf("unable to set write deadline (%w)", err)
	}

	if _, err := conn.WriteToUDP(data, peer); err != nil {
		return err
	}

	return nil
}

func readMessage(conn *net.UDPConn) (*message, *net.UDPAddr, error) {
	data := make([]byte, udp.MaxPacketSize)
	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return nil, nil, fmt.Errorf("error setting read deadline (%w)", err)
	}
	if n, peer, err := conn.ReadFromUDP(data); err == nil {
		if m, err := decodeMessage(data[:n]); err == nil {
			return m, peer, nil
		} else {
			return nil, nil, fmt.Errorf("error decoding message from [%s] (%w)", peer, err)
		}
	} else {
		return nil, nil, err
	}
}

func handleHello(m *message, conn *net.UDPConn, peer *net.UDPAddr, handler HelloHandler) error {
	if m != nil {
		switch m.messageType {
		case Hello:
			if m.sequence != -1 {
				return fmt.Errorf("hello expects sequence -1 [%s]", peer)
			}
			if m.fragment != 0 || m.ofFragments != 1 {
				return fmt.Errorf("hello expects single fragment [%s]", peer)
			}
			linkId := &identity.TokenId{Token: string(m.payload)}
			handler.HandleHello(linkId, conn, peer)

			return nil

		default:
			return fmt.Errorf("expected hello, not [%d] from [%s]", m.messageType, peer)
		}
	} else {
		return fmt.Errorf("nil message")
	}
}

func handleMessage(m *message, conn *net.UDPConn, peer *net.UDPAddr, handler MessageHandler) error {
	if m.fragment != 0 || m.ofFragments != 1 {
		return fmt.Errorf("ping expects single fragment [%s]", peer)
	}

	switch m.messageType {
	case Ping:
		replyFor, err := readInt32(m.payload)
		if err != nil {
			return fmt.Errorf("ping expects replyFor in payload [%s] (%w)", peer, err)
		}
		handler.HandlePing(m.sequence, replyFor, conn, peer)

		return nil

	case Payload:
		p, err := decodePayload(m)
		if err != nil {
			return fmt.Errorf("error decoding payload for peer [%s] (%w)", peer, err)
		}
		handler.HandlePayload(p, m.sequence, conn, peer)

		return nil

	case Acknowledgement:
		a, err := decodeAcknowledgement(m)
		if err != nil {
			return fmt.Errorf("error decoding acknowledgement for peer [%s] (%w)", peer, err)
		}
		handler.HandleAcknowledgement(a, m.sequence, conn, peer)

		return nil

	case WindowReport:
		lowWater, highWater, gaps, count, err := decodeWindowReport(m)
		if err != nil {
			return fmt.Errorf("error decoding window report for peer [%s] (%w)", peer, err)
		}
		handler.HandleWindowReport(lowWater, highWater, gaps, count, conn, peer)

		return nil

	case WindowSizeRequest:
		newWindowSize, err := decodeWindowSizeRequest(m)
		if err != nil {
			return fmt.Errorf("error decoding window size request for peer [%s] (%w)", peer, err)
		}
		handler.HandleWindowSizeRequest(newWindowSize, conn, peer)

		return nil

	default:
		return fmt.Errorf("unexpected message type [%d] from [%s]", m.messageType, peer)
	}
}
