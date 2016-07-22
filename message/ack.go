package message

import "fmt"

type AckMessage struct {
	header

	returnCode     AckCode
}

var _ Message = (*AckMessage)(nil)

func NewAckMessage() *AckMessage {
	msg := &AckMessage{}
	msg.SetType(ACK)

	return msg
}

func (this AckMessage) String() string {
	return fmt.Sprintf("%s, Return code=%q\n", this.header, this.returnCode)
}

func (this *AckMessage) ReturnCode() AckCode {
	return this.returnCode
}

func (this *AckMessage) SetReturnCode(ret AckCode) {
	this.returnCode = ret
}

func (this *AckMessage) Len() int {

	ml := this.msglen()

	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return this.header.msglen() + ml
}

func (this *AckMessage) Decode(src []byte) (int, error) {
	total := 0

	n, err := this.header.decode(src)
	total += n
	if err != nil {
		return total, err
	}

	if this.header.mtype != byte(ACK) {
                return total, ErrType
        }

	b := src[total]

	this.returnCode = AckCode(b)
	total++

	return total, nil
}

func (this *AckMessage) Encode(dst []byte) (int, error) {
	ml := this.msglen()


	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0, err
	}

	total := 0

	n, err := this.header.encode(dst[total:])
	total += n
	if err != nil {
		return 0, err
	}


	dst[total] = this.returnCode.Value()
	total++

	return total, nil
}

func (this *AckMessage) msglen() int {
	return 1
}
