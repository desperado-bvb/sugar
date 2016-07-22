package message

import (
	"fmt"
)

type StopMessage struct {
	header

	payload []byte
}

var _ Message = (*StopMessage)(nil)

func NewStopMessage() *StopMessage {
	msg := &StopMessage{}
	msg.SetType(STOP)

	return msg
}

func (this StopMessage) String() string {
	return fmt.Sprintf("%s, Payload=%v", this.header,  this.payload)
}

func (this *StopMessage) Payload() []byte {
	return this.payload
}

func (this *StopMessage) SetPayload(v []byte) {
	this.payload = v
}

func (this *StopMessage) Len() int {

	ml := this.msglen()

	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return this.header.msglen() + ml
}

func (this *StopMessage) Decode(src []byte) (int, error) {
	total := 0

	hn, err := this.header.decode(src[total:])
	total += hn
	if err != nil {
		return total, err
	}

	
	n := int(this.RemainingLength())
        this.payload = src[total :total+n]
	total += len(this.payload)


	return total, nil
}

func (this *StopMessage) Encode(dst []byte) (int, error) {

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

func (this *StopMessage) msglen() int {
	total := len(this.payload)

	return total
}
