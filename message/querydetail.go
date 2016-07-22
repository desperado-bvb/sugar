package message

import (
	"fmt"
)

type QuerydetailMessage struct {
	header

	payload []byte
}

var _ Message = (*QuerydetailMessage)(nil)

func NewQuerydetailMessage() *QuerydetailMessage {
	msg := &QuerydetailMessage{}
	msg.SetType(QUERYDETAIL)

	return msg
}

func (this QuerydetailMessage) String() string {
	return fmt.Sprintf("%s, Payload=%v", this.header,  this.payload)
}

func (this *QuerydetailMessage) Payload() []byte {
	return this.payload
}

func (this *QuerydetailMessage) SetPayload(v []byte) {
	this.payload = v
}

func (this *QuerydetailMessage) Len() int {

	ml := this.msglen()

	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return this.header.msglen() + ml
}

func (this *QuerydetailMessage) Decode(src []byte) (int, error) {
	total := 0

	hn, err := this.header.decode(src[total:])
	total += hn
	if err != nil {
		return total, err
	}

	if this.header.mtype != byte(QUERYDETAIL) {
		return total, ErrType
	}

	n := int(this.RemainingLength())
	this.payload = src[total :total+n]
	total += len(this.payload)


	return total, nil
}

func (this *QuerydetailMessage) Encode(dst []byte) (int, error) {

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

	fmt.Println(dst)

	return total, nil
}

func (this *QuerydetailMessage) msglen() int {
	total := len(this.payload)

	return total
}
