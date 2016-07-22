package agent

import (
	"errors"
	"io"
	"fmt"
	"time"
	"encoding/binary"
	"strings"
	"syscall"
	"encoding/json"
	

	"github.com/desperado-bvb/sugar/message"
	"github.com/golang/glog"
)

var (
	errDisconnect = errors.New("Disconnect")
)

func (this *Server) processor() {
	defer func() {
		if r := recover(); r != nil {
			glog.Errorf("(%s) Recovering from panic: %v", this.ID, r)
		}

		this.wgStopped.Done()
		this.Close()

	}()


	this.wgStarted.Done()

	for {
		mtype, total, err := this.peekMessageSize()
		if err != nil {
			if err != io.EOF {
			    glog.Errorf("(%s) Error peeking next message size: %v", this.ID, err)
			}
			return
		}

		msg, _, err := this.peekMessage(mtype, total)
		if err != nil {
			glog.Errorf("(%s) Error peeking next message: %v", this.ID, err)
			return
		}

		err = this.processIncoming(msg)
		if err != nil {
			if err != errDisconnect {
				glog.Errorf("(%s) Error processing %s: %v", this.ID, msg.Name(), err)
			} else {
				return
			}
		}

		_, err = this.in.ReadCommit(total)
		if err != nil {
			if err != io.EOF {
				glog.Errorf("(%s) Error committing %d read bytes: %v", this.ID, total, err)
			}
			return
		}

		if !this.running && this.in.Len() == 0 {
			return
		}
	}
}

func (this *Server) processIncoming(msg message.Message) error {
        var err error = nil
	fmt.Println(msg)
        switch msg := msg.(type) {
        case *message.StartMessage:

                resp := this.processStartProcess(msg)
               // _, err = this.writeMessage(resp)
		err = writeMessage(this.conn, resp)
		fmt.Println(resp)

	case *message.StopMessage:
		resp := this.processStopProcess(msg)
		//_, err = this.writeMessage(resp)
		err = writeMessage(this.conn, resp)
		fmt.Println(resp)

	case *message.QueryMessage:
		resp := this.processQuery()
		//_, err = this.writeMessage(resp)
		err = writeMessage(this.conn, resp)
                fmt.Println(resp)

        default:
                return fmt.Errorf("(%s) invalid message type %s.", this.ID, msg.Name())

        }

        return err
}

func (this *Server) processStartProcess(msg *message.StartMessage) *message.AckMessage{
	c := string(msg.Payload())
	cmd := strings.Split(c, " ")

	resp := message.NewAckMessage()
	length := len(cmd)
	if length == 0 {
                resp.SetReturnCode(message.ErrStart)
		return resp
	}

	execSpec := &syscall.ProcAttr{
        }


        fork, _ := syscall.ForkExec(cmd[0], cmd[1:], execSpec)
	if fork == 0 {
		resp.SetReturnCode(message.ErrStart)
		return resp
	}

	spid := fmt.Sprintf("%d", fork)

	this.mu.Lock()
	this.process[spid] = ProcessInfo{Cmd:c, Start:time.Now(), Pid:fork}
	this.mu.Unlock()


	resp.SetReturnCode(message.StartSuccess)

        return resp
}

func (this *Server) processStopProcess(msg *message.StopMessage) *message.AckMessage {

	resp := message.NewAckMessage()
	pid, _ := binary.Varint(msg.Payload())
	if pid < 1 {
		resp.SetReturnCode(message.ErrStop)
                return resp
	}

	spid := fmt.Sprintf("%d", pid)

	_, ok := this.process[spid]
	if !ok {
		resp.SetReturnCode(message.ErrStop)
                return resp
	}

	err := syscall.Kill(int(pid), syscall.SIGKILL)
	if err != nil {
		resp.SetReturnCode(message.ErrStop)
                return resp
	}

	this.mu.Lock()
	delete(this.process, spid)
	this.mu.Unlock()

	resp.SetReturnCode(message.StopSuccess)
	return resp
	
}

func (this *Server) processQuery() *message.QuerydetailMessage {
	resp := message.NewQuerydetailMessage()

	x, err := json.Marshal(this.process)
	fmt.Println(x, err)
	if err != nil {
		return resp
	}
	
	resp.SetPayload(x)

	return resp
}
