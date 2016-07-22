package agent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"net"
	"sync"
	"syscall"
	"strconv"
	"io/ioutil"
	"time"
	"github.com/desperado-bvb/sugar/message"
	"github.com/golang/glog"
)


var (
	ErrInvalidConnectionType  error = errors.New("service: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("service: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("service: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("service: buffer has insufficient data.")
)

type ProcessInfo struct {
	Cmd     string    	`json:"command"`
	Start   time.Time	`json:"start"`
	Pid     int		`json:"pid"`
}


type Server struct {
	ID		string
	done chan	bool
	process         map[string] ProcessInfo
	mu		sync.RWMutex
	wmu 		sync.Mutex
	running		bool
	start		time.Time
	conn    	io.Closer

	wgStarted sync.WaitGroup
	wgStopped sync.WaitGroup

	in *buffer
	out *buffer

	intmp  []byte
        outtmp []byte

	opts		*Options
}

func New(opts *Options) *Server {
	processOptions(opts)

	this := &Server{
		ID:	genID(),
		opts:	opts,
		process: make(map[string] ProcessInfo),
		start:	time.Now(),
	}


	return this
}

func (this *Server) register() {
	request := message.NewRegisterMessage2(this.ID, this.opts.Cluster, this.opts.DataCenter, this.opts.KeepAlive)
	if err := writeMessage(this.conn, request); err != nil {
                glog.Errorf("server/slave: could not send register message; %v", err)
                os.Exit(1)
        }

        _, err := getAckMessage(this.conn)
        if err != nil {
		glog.Errorf("server/slave: fail to register; %v", err)
		os.Exit(1)
        }
}

func (this *Server) handleSignals() {
	if this.opts.NoSigs == 0 {
		return
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			glog.Infof("Server Exiting..")
			os.Exit(0)
		}
	}()
}

func (this *Server) logPid() {
        pidStr := strconv.Itoa(os.Getpid())
        err := ioutil.WriteFile(this.opts.PidFile, []byte(pidStr), 0660)
        if err != nil {
                fmt.Errorf("Could not write pidfile: %v\n", err)
                os.Exit(1)
        }
}

func (this *Server) isRunning() bool {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.running
}

func (this *Server) Close() {

	this.mu.Lock()

	if !this.running  {
		this.mu.Unlock()
	}

	if this.conn != nil {
        	this.conn.Close()
        }

	this.running = false

	var process []int

	for _, p := range this.process {
		process = append(process, p.Pid)
	}

	this.mu.Unlock()

	for _, pid := range process {
		syscall.Kill(pid, syscall.SIGKILL)
	}
}

func (this *Server) Start() {
	conn, err := net.Dial("tcp", this.opts.Host)
        if err != nil {
                glog.Infof("server/slave: could not connect master service; %v", err)
                os.Exit(1)
        }

        this.conn = conn

	this.mu.Lock()

	if this.running {
		return
	}

        this.register()


        glog.Infof("server/slave: server is ready...")
        this.handleSignals()
        this.running = true
        if this.opts.PidFile != "" {
                this.logPid()
        }

	this.mu.Unlock()

	this.in, err = newBuffer(defaultBufferSize)
	if err != nil {
		os.Exit(1)
	}

	this.out, err = newBuffer(defaultBufferSize)
	if err != nil {
		os.Exit(1)
	}


	this.wgStarted.Add(1)
	this.wgStopped.Add(1)
	go this.processor()

	this.wgStarted.Add(1)
	this.wgStopped.Add(1)
	go this.receiver()

	this.wgStarted.Add(1)
	this.wgStopped.Add(1)
	go this.sender()

	this.wgStarted.Wait()

	this.handleChild()

}

func (this *Server) handleChild() {
	var ws syscall.WaitStatus
        var usage = syscall.Rusage{}

	tmpDelay := ACCEPT_MIN_SLEEP
	for {
        	if len(this.process) == 0 {
			time.Sleep(tmpDelay)
                        tmpDelay *= 2
                        if tmpDelay > ACCEPT_MAX_SLEEP {
                            tmpDelay = ACCEPT_MAX_SLEEP
                        }

			continue
		}

		wpid, _ := syscall.Wait4(-1, &ws, 0, &usage)

		if wpid == -1 {
			this.process = make(map[string] ProcessInfo)	
		}

		spid := fmt.Sprintf("%d", wpid)
		this.mu.Lock()
        	delete(this.process, spid)
        	this.mu.Unlock()
	}	
}
