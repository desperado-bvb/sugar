package message

import (
	"encoding/binary"
	"fmt"
)

type RegisterMessage struct {
	header

	keepAlive uint16
	name []byte
}

var _ Message = (*RegisterMessage)(nil)

func NewRegisterMessage() *RegisterMessage {
	msg := &RegisterMessage{
        }
        msg.SetType(REGISTER)

        return msg
}

func NewRegisterMessage2(ID string, cluster string, dataCenter string, keepAlive uint16) *RegisterMessage {
	msg := &RegisterMessage{
		keepAlive: keepAlive,
		name: []byte(cluster+"/"+dataCenter+"/"+ID),
	}
	msg.SetType(REGISTER)

	return msg
}

func (this RegisterMessage) String() string {
	return fmt.Sprintf("%s, Connect KeepAlive=%d, Name=%v",
		this.header,
		this.KeepAlive(),
		string(this.ClusterName()),
	)
}

func (this *RegisterMessage) KeepAlive() uint16 {
	return this.keepAlive
}

func (this *RegisterMessage) SetKeepAlive(v uint16) {
	this.keepAlive = v
}

func (this *RegisterMessage) ClusterName() []byte {
	return this.name
}

func (this *RegisterMessage) SetClusterName(v []byte) error {

	this.name = v

	return nil
}

func (this *RegisterMessage) Len() int {

	ml := this.msglen()

	if err := this.SetRemainingLength(int32(ml)); err != nil {
		return 0
	}

	return this.header.msglen() + ml
}

func (this *RegisterMessage) Decode(src []byte) (int, error) {
	total := 0

	n, err := this.header.decode(src[total:])
	if err != nil {
		return total + n, err
	}
	total += n

	if this.header.mtype != byte(REGISTER) {
		return total, ErrType
	}

	this.keepAlive = binary.BigEndian.Uint16(src[total:])
        total += 2

	n = int(this.RemainingLength())
        this.name = src[total :total+n-2]
        total += len(this.name)	

	return total, nil
}

func (this *RegisterMessage) Encode(dst []byte) (int, error) {
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

	binary.BigEndian.PutUint16(dst[total:], this.keepAlive)
        total += 2

	copy(dst[total:], this.name)
        total += len(this.name)	

	return total, nil
}

func (this *RegisterMessage) msglen() int {
	return 2+len(this.name)
}
