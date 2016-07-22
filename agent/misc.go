package agent

import (
	"encoding/binary"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"

	"github.com/desperado-bvb/sugar/message"
)

func getRegistertMessage(conn io.Closer) (*message.RegisterMessage, error) {
	buf, err := getMessageBuffer(conn)
	if err != nil {
		//glog.Debugf("Receive error: %v", err)
		return nil, err
	}

	msg := message.NewRegisterMessage()

	_, err = msg.Decode(buf)
	return msg, err
}

func getAckMessage(conn io.Closer) (*message.AckMessage, error) {
	buf, err := getMessageBuffer(conn)

	msg := message.NewAckMessage()

	_, err = msg.Decode(buf)
	return msg, err
}

func writeMessage(conn io.Closer, msg message.Message) error {
	buf := make([]byte, msg.Len())
	_, err := msg.Encode(buf)
	if err != nil {
		//glog.Debugf("Write error: %v", err)
		return err
	}

	return writeMessageBuffer(conn, buf)
}

func getMessageBuffer(c io.Closer) ([]byte, error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}

	conn, ok := c.(net.Conn)
	if !ok {
		return nil, ErrInvalidConnectionType
	}

	var (
		buf []byte

		b []byte = make([]byte, 1)

		l int = 0
	)

	for {
		if l > 5 {
			return nil, fmt.Errorf("connect/getMessage: 4th byte of remaining length has continuation bit set")
		}

		n, err := conn.Read(b[0:])
		if err != nil {
			//glog.Debugf("Read error: %v", err)
			return nil, err
		}

		// Technically i don't think we will ever get here
		if n == 0 {
			continue
		}

		buf = append(buf, b...)
		l += n

		if l >= 5 {
			break
		}
	}

	remlen, _ := binary.Uvarint(buf[1:5])
	buf = append(buf, make([]byte, remlen)...)

	for l < len(buf) {
		n, err := conn.Read(buf[l:])
		if err != nil {
			return nil, err
		}
		l += n
	}

	return buf, nil
}

func writeMessageBuffer(c io.Closer, b []byte) error {
	if c == nil {
		return ErrInvalidConnectionType
	}

	conn, ok := c.(net.Conn)
	if !ok {
		return ErrInvalidConnectionType
	}

	_, err := conn.Write(b)
	return err
}

// Copied from http://golang.org/src/pkg/net/timeout_test.go
func isTimeout(err error) bool {
	e, ok := err.(net.Error)
	return ok && e.Timeout()
}

func genID() string {
	u := make([]byte, 16)
	io.ReadFull(rand.Reader, u)
	return hex.EncodeToString(u)
}
