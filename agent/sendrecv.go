package agent

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/desperado-bvb/sugar/message"
	"github.com/golang/glog"
)

func (this *Server) receiver() {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", this.ID, r)
		}

		this.wgStopped.Done()

	}()


	this.wgStarted.Done()
	for {

		conn := this.conn.(net.Conn)
		_, err := this.in.ReadFrom(conn)

		if err != nil {
			if err != io.EOF {
				glog.Errorf("(%s) error reading from connection: %v", this.ID, err)
			}
			return
		}
	}

}

func (this *Server) sender() {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", this.ID, r)
		}

		this.wgStopped.Done()

	}()


	this.wgStarted.Done()

	for {
		conn := this.conn.(net.Conn)
		_, err := this.out.WriteTo(conn)

		if err != nil {
			if err != io.EOF {
				glog.Errorf("(%s) error writing data: %v", this.ID, err)
			}
			return
		}
	}
}

func (this *Server) peekMessageSize() (message.MessageType, int, error) {
	var (
		b   []byte
		err error
		cnt int = 2
	)

	if this.in == nil {
		err = ErrBufferNotReady
		return 0, 0, err
	}

	for {
		if cnt > 5 {
			return 0, 0, fmt.Errorf("sendrecv/peekMessageSize: 4th byte of remaining length has continuation bit set")
		}

		b, err = this.in.ReadWait(cnt)
		if err != nil {
			return 0, 0, err
		}

		if len(b) < cnt {
			continue
		}

		if cnt < 5 {
			cnt++
		} else {
			break
		}
	}

	remlen, _ := binary.Uvarint(b[1:])

	total := int(remlen) + 5

	mtype := message.MessageType(b[0])

	return mtype, total, err
}

func (this *Server) peekMessage(mtype message.MessageType, total int) (message.Message, int, error) {
	var (
		b    []byte
		err  error
		i, n int
		msg  message.Message
	)

	if this.in == nil {
		return nil, 0, ErrBufferNotReady
	}

	for i = 0; ; i++ {
		b, err = this.in.ReadWait(total)
		if err != nil && err != ErrBufferInsufficientData {
			return nil, 0, err
		}

		if len(b) >= total {
			break
		}
	}

	msg, err = mtype.New()
	if err != nil {
		return nil, 0, err
	}

	n, err = msg.Decode(b)
	return msg, n, err
}

func (this *Server) readMessage(mtype message.MessageType, total int) (message.Message, int, error) {
	var (
		b   []byte
		err error
		n   int
		msg message.Message
	)

	if this.in == nil {
		err = ErrBufferNotReady
		return nil, 0, err
	}

	if len(this.intmp) < total {
		this.intmp = make([]byte, total)
	}

	l := 0
	for l < total {
		n, err = this.in.Read(this.intmp[l:])
		l += n
		if err != nil {
			return nil, 0, err
		}
	}

	b = this.intmp[:total]

	msg, err = mtype.New()
	if err != nil {
		return msg, 0, err
	}

	n, err = msg.Decode(b)
	return msg, n, err
}

func (this *Server) writeMessage(msg message.Message) (int, error) {
	var (
		l    int = msg.Len()
		m, n int
		err  error
		buf  []byte
		wrap bool
	)

	if this.out == nil {
		return 0, ErrBufferNotReady
	}

	this.wmu.Lock()
	defer this.wmu.Unlock()

	buf, wrap, err = this.out.WriteWait(l)
	if err != nil {
		return 0, err
	}

	if wrap {
		if len(this.outtmp) < l {
			this.outtmp = make([]byte, l)
		}

		n, err = msg.Encode(this.outtmp[0:])
		if err != nil {
			return 0, err
		}

		m, err = this.out.Write(this.outtmp[0:n])
		if err != nil {
			return m, err
		}
	} else {
		n, err = msg.Encode(buf[0:])
		if err != nil {
			return 0, err
		}


		m, err = this.out.WriteCommit(n)
		if err != nil {
			return 0, err
		}
	}

	return m, nil
}
