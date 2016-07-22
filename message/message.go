package message

import (
	"fmt"
	"errors"
)

var (
	ErrType error = errors.New("msgtype/NewMessage: Invalid message type")
)

type Message interface {
	Name() string

	Type() MessageType

	Encode([]byte) (int, error)

	Decode([]byte) (int, error)

	Len() int
}

type MessageType byte

const (
	REGISTER = iota

	ACK

	START

	STOP

	QUERY

	HEARTBEAT

	QUERYDETAIL
)

func (this MessageType) String() string {
	return this.Name()
}

func (this MessageType) Name() string {
	switch this {
	case REGISTER:
		return "REGISTER"
	case ACK:
		return "ACK"
	case START:
		return "START"
	case STOP:
		return "STOP"
	case QUERY:
		return "QUERY"
	case HEARTBEAT:
		return "HEARTBEAT"
	case QUERYDETAIL:
		return "QUERYDETAIL"
	}

	return "UNKNOWN"
}

func (this MessageType) New() (Message, error) {
	switch this {
	case REGISTER:
		return NewRegisterMessage(), nil
	case ACK:
		return NewAckMessage(), nil
	case START:
		return NewStartMessage(), nil
	case STOP:
		return NewStopMessage(), nil
	case QUERY:
		return NewQueryMessage(), nil
	case HEARTBEAT:
		return NewHeartbeatMessage(), nil
	case QUERYDETAIL:
		return NewQuerydetailMessage(), nil
	}

	return nil, fmt.Errorf("msgtype/NewMessage: Invalid message type %d", this)
}
