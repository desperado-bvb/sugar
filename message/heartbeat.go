package message

import (
	"fmt"
)

type HeartbeatMessage struct {
	header

	payload []byte
}

var _ Message = (*HeartbeatMessage)(nil)

func NewHeartbeatMessage() *HeartbeatMessage {
	msg := &HeartbeatMessage{}
	msg.SetType(HEARTBEAT)

	return msg
}

func (this HeartbeatMessage) String() string {
	return fmt.Sprintf("%s, Payload=%v", this.header,  this.payload)
}

func (this *HeartbeatMessage) Payload() []byte {
	return this.payload
}

func (this *HeartbeatMessage) SetPayload(v []byte) {
	this.payload = v
}

func (this *HeartbeatMessage) Len() int {

	ml := this.msglen()

	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return this.header.msglen() + ml
}

func (this *HeartbeatMessage) Decode(src []byte) (int, error) {
	total := 0

	hn, err := this.header.decode(src[total:])
	total += hn
	if err != nil {
		return total, err
	}

	n := this.RemainingLength()
        this.payload = src[total :total+int(n)]
	total += len(this.payload)


	return total, nil
}

func (this *HeartbeatMessage) Encode(dst []byte) (int, error) {

	ml := this.msglen()

	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0, err
	}

	total := 0

	n, err := this.header.encode(dst[total:])
	total += n
	if err != nil {
		return total, err
	}

	copy(dst[total:], this.payload)
	total += len(this.payload)

	return total, nil
}

func (this *HeartbeatMessage) msglen() int {
	total := len(this.payload)

	return total
}
