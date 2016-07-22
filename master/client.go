/*
the client thet can handle the request , like startProcess, stopProcess, QueryProcesss and so on
*/

package master

import (
	"fmt"
	"io"
	"sync"
	"encoding/binary"
	"sync/atomic"

	"github.com/desperado-bvb/sugar/message"
	"github.com/golang/glog"
)


type client struct {
	id uint64

	name string

	keepAlive int

	conn io.Closer


	wgStarted sync.WaitGroup
	wgStopped sync.WaitGroup

	wmu sync.Mutex

	mu sync.Mutex

	intmp  []byte
	outtmp []byte

	closed int64

	done chan struct{}

	in *buffer
	out *buffer


	srv	*Server


	deleteCallback func(*client)
	deleter        sync.Once

    	ip string
}

func (this *client) start() error {
	var err error

	this.in, err = newBuffer(defaultBufferSize)
	if err != nil {
		return err
	}

	this.out, err = newBuffer(defaultBufferSize)
	if err != nil {
		return err
	}

	return nil
}

func (this *client) StartProcess(cmd string) error {
	request := message.NewStartMessage()
	request.SetPayload([]byte(cmd))

	if err := writeMessage(this.conn, request); err != nil {
		glog.Errorf("server/client: could not send start message; %v", err)
		return  err
	}

	rep, err := getAckMessage(this.conn)
	if err != nil {
                glog.Errorf("server/client: fail to start process; %v", err)
                return err
        }

	fmt.Println(rep)

	if rep.ReturnCode() == message.StartSuccess {
		return nil
	}

	return ErrStartProcess
}

func (this *client) StopProcess(pid int) error {
	request := message.NewStopMessage()
	var b = make([]byte,4,4)
	binary.PutVarint(b, int64(pid))
	request.SetPayload(b)
        if err := writeMessage(this.conn, request); err != nil {
                glog.Errorf("server/client: could not send stop message; %v", err)
                return err
        }

        rep, err := getAckMessage(this.conn)
        if err != nil {
                glog.Errorf("server/client: fail to stop process; %v", err)
                return err
        }

	fmt.Println(rep)

	if rep.ReturnCode() == message.StopSuccess {
                return nil
        }       
        
        return ErrStopProcess
}

func (this *client) QueryProcess() ([]byte, error) {
	request := message.NewQueryMessage()
        if err := writeMessage(this.conn, request); err != nil {
                glog.Errorf("server/client: could not send query  message; %v", err)
                return nil, err
        }       
        
        rep, err := getQuerydetailMessage(this.conn)
        if err != nil {
                glog.Errorf("server/client: fail to query process; %v", err)
                return nil, err
        }       

	fmt.Println(rep)

	return rep.Payload(), nil

}

func (this *client) stop() {
	if atomic.LoadInt64(&this.closed) == 1 {
        glog.Infof("client id %v is dead in ip %v", this.cid(), this.ip)
		return
	}

	this.close()

	this.deleter.Do(func() { this.deleteCallback(this) })
}

func (this *client)close() {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", this.cid(), r)
		}
	}()

	doit := atomic.CompareAndSwapInt64(&this.closed, 0, 1)
	if !doit {
		return
	}

	if this.done != nil {
		close(this.done)
	}

	if this.conn != nil {
		this.conn.Close()
	}

	if this.in != nil {
		this.in.Close()
	}

	if this.out != nil {
		this.out.Close()
	}

	this.wgStopped.Wait()

	this.conn = nil
	this.in = nil
	this.out = nil

    glog.Infof("ip:%v cid:%v closed", this.ip, this.cid())
}

func (this *client) isDone() bool {
	select {
	case <-this.done:
		return true

	default:
	}

	return false
}

func (this *client) cid() string {
	return fmt.Sprintf("%d", this.id)
}
