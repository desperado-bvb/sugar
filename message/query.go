package message

import (
	"fmt"
)

type QueryMessage struct {
	header
}

var _ Message = (*QueryMessage)(nil)

func NewQueryMessage() *QueryMessage {
	msg := &QueryMessage{}
	msg.SetType(QUERY)

	return msg
}

func (this QueryMessage) String() string {
	return fmt.Sprintf("%s", this.header)
}

func (this *QueryMessage) Len() int {

	if err := this.SetRemainingLength(0); err != nil {
		return 0
	}

	return this.header.msglen()
}

func (this *QueryMessage) Decode(src []byte) (int, error) {
	total := 0

	hn, err := this.header.decode(src[total:])
	total += hn
	if err != nil {
		return total, err
	}

	return total, nil
}

func (this *QueryMessage) Encode(dst []byte) (int, error) {

	if err := this.SetRemainingLength(0); err != nil {
		return 0, err
	}

	total := 0

	n, err := this.header.encode(dst[total:])
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (this *QueryMessage) msglen() int {
	return 0
}
