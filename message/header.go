package message

import (
	"encoding/binary"
	"fmt"
)


type header struct {
	remlen int32

	mtype  byte
}

func (this header) String() string {
	return fmt.Sprintf("Type=%q, Remaining Length=%d", this.Type().Name(),  this.remlen)
}

func (this *header) Name() string {
	return this.Type().Name()
}


func (this *header) Type() MessageType {

	return MessageType(this.mtype)
}

func (this *header) SetType(mtype MessageType) error {

	this.mtype = byte(mtype)

	return nil
}

func (this *header) RemainingLength() int32 {
	return this.remlen
}

func (this *header) SetRemainingLength(remlen int32) error {

	this.remlen = remlen

	return nil
}

func (this *header) Len() int {
	return this.msglen()
}


func (this *header) encode(dst []byte) (int, error) {

	total := 0

	dst[total] = this.mtype
	total += 1

	n := binary.PutUvarint(dst[total:], uint64(this.remlen))
	total += n

	return total, nil
}

func (this *header) decode(src []byte) (int, error) {
	total := 0

	this.mtype = src[total]

	total++

	remlen, m := binary.Uvarint(src[total:])
	total += m
	this.remlen = int32(remlen)


	return total, nil
}

func (this *header) msglen() int {
	return 5
}
