package marlin

import (
	"encoding/json"
	"errors"
	"net"
)

const (
	Ping    = "Ping"
	Ack     = "Ack"
	Alive   = "Alive"
	Suspect = "Suspect"
	Confirm = "Confirm"
)

var InvalidMessage = errors.New("invalid message")

type Message struct {
	Type        string `json:"type"`
	From        string `json:"from"`
	Who         string `json:"who"`
	fromAddress *net.UDPAddr
	whoAddress  *net.UDPAddr
}

func NewMessage(message []byte) (*Message, error) {
	msg := Message{}
	if err := json.Unmarshal(message, &msg); err != nil {
		return nil, InvalidMessage
	} else {
		return &msg, nil
	}
}

func (m *Message) FromAddress() (*net.UDPAddr, error) {
	if m.fromAddress != nil {
		return m.fromAddress, nil
	}

	var err error
	if m.fromAddress, err = net.ResolveUDPAddr("udp4", m.From); err != nil {
		return nil, InvalidMessage
	} else {
		return m.fromAddress, nil
	}
}

func (m *Message) WhoAddress() (*net.UDPAddr, error) {
	if m.whoAddress != nil {
		return m.whoAddress, nil
	}

	var err error
	if m.whoAddress, err = net.ResolveUDPAddr("udp4", m.Who); err != nil {
		return nil, InvalidMessage
	} else {
		return m.whoAddress, nil
	}
}
