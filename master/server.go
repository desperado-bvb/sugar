package master

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"os"
	"os/signal"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"strconv"
	"io/ioutil"
	"time"
	"github.com/desperado-bvb/sugar/message"
	"github.com/golang/glog"
)

const (
    OFFLINE byte = '0'
    ONLINE byte = '1'
)

var (
	ErrInvalidConnectionType  error = errors.New("service: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("service: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("service: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("service: buffer has insufficient data.")
	ErrStartProcess           error = errors.New("service: fail to start process.")
	ErrStopProcess            error = errors.New("service: fail to stop  process.")
)

type ProcessInfo struct {
        Cmd     string          `json:"command"`
        Start   time.Time       `json:"start"`
        Pid     int             `json:"pid"`
}


type Server struct {
	ID		string
	done		chan bool
	gcid		uint64

	tcpListener     net.Listener
	httpListener    net.Listener

	clients         map[string] *client
	waitGroup       sync.WaitGroup

	mu		sync.RWMutex
	running		bool
	start		time.Time

	opts		*Options
}

func New(opts *Options) *Server {
	processOptions(opts)

	this := &Server{
		ID:		genID(),
		opts:	opts,
		done:	make(chan bool, 1),
		start:	time.Now(),
	}

	this.mu.Lock()
	defer this.mu.Unlock()

	this.clients = make(map[string]*client)

	this.handleSignals()

	return this
}

func (s *Server) handleSignals() {
	if s.opts.NoSigs == 0 {
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

func (s *Server) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (this *Server) Start() {
	glog.Infof("server/ListenAndServe: server is ready...")
	this.running = true


	if this.opts.PidFile != "" {
		this.logPid()
	}

	httpAddr, err := net.ResolveTCPAddr("tcp", this.opts.HTTPAddress)
	if err != nil {
		glog.Errorf("server/http: Could notresolve Http addr: %v\n", err)
		os.Exit(1)
	}

	httpListener, err := net.Listen("tcp", httpAddr.String())
        if err != nil {
                glog.Errorf("server/http: listen (%s) failed - %s", httpAddr, err)
                os.Exit(1)
        }
        this.httpListener = httpListener


        this.waitGroup.Add(1)
        go this.httpServer()



	if this.opts.Port != 0 {
		this.ListenAndServer()
	}
}

func (this *Server) httpServer() {
    defer this.waitGroup.Done()

    handler := &httpServer{
                Service: this,
    }

    hs := &http.Server{
        Handler: handler,
    }

    err := hs.Serve(this.httpListener)
    if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
        glog.Errorf("server/http: http.Serve() - %s", err)
    }

    glog.Infof("server/http: closing %s", this.httpListener.Addr())
}

func (this *Server) ListenAndServer() {
	url := fmt.Sprintf(":%d", this.opts.Port)
	ln, err := net.Listen("tcp", url)
	if err != nil {
		glog.Fatalf("Error listening on port: %s, %q", url, err)
		return
	}

	this.mu.Lock()
	this.tcpListener = ln
	this.mu.Unlock()

	glog.Infof("server/ListenAndServe: TCP server is ready...")

	tmpDelay := ACCEPT_MIN_SLEEP

	for this.isRunning() {
		conn, err := ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(tmpDelay)
				tmpDelay *= 2
				if tmpDelay > ACCEPT_MAX_SLEEP {
					tmpDelay = ACCEPT_MAX_SLEEP
				}
			} else {
				glog.Infof("Accept error: %v", err)
			}
			continue
		}
		tmpDelay = ACCEPT_MIN_SLEEP
		go this.createClient(conn)
	}

	glog.Infof("Server Exiting..")
	this.done <- true

}

func (this *Server) Close() {

	this.mu.Lock()

	if !this.running  {
		this.mu.Unlock()
	}

	this.running = false

	var conns []*client

	for _, c := range this.clients {
		conns = append(conns, c)
	}

	if this.tcpListener != nil {
		this.tcpListener.Close()
		this.tcpListener = nil
	}

	if this.httpListener != nil {
		this.httpListener.Close()
		this.httpListener = nil
	}

	this.mu.Unlock()

	this.waitGroup.Wait()

	for _, c := range conns {
		c.stop()
	}
}

func (this *Server) createClient(c io.Closer) (svc *client, err error) {
	if c == nil {
		return nil, ErrInvalidConnectionType
	}

	defer func() {
		if err != nil {
			glog.Errorf("handle connect: %s", err)
			c.Close()
		}
	}()

	conn, ok := c.(net.Conn)
	if !ok {
		return nil, ErrInvalidConnectionType
	}

	//conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(this.opts.ConnectTimeout)))

	resp := message.NewAckMessage()

	req, err := getRegistertMessage(conn)
	if err != nil {
		if cerr, ok := err.(message.AckCode); ok {
			resp.SetReturnCode(cerr)
			writeMessage(conn, resp)
		}
		return nil, err
	}

	if req.KeepAlive() == 0 {
		req.SetKeepAlive(60 * 60 * 2)
	}

	deleteCallback := func(s *client) {
		this.DeleteExistingSvc(s)
	}

	svc = &client{
		id:		atomic.AddUint64(&this.gcid, 1),
		srv:		this,
		name:	        string(req.ClusterName()),
		keepAlive:      int(req.KeepAlive()),

		conn:		conn,
		deleteCallback: deleteCallback,
        	ip:         conn.RemoteAddr().String(),
	}

	resp.SetReturnCode(message.ConnectionAccepted)


	if err = writeMessage(c, resp); err != nil {
		return nil, err
	}


	this.mu.RLock()
	_, ok = this.clients[svc.name]
	this.mu.RUnlock()

	if ok {
		return
	}

	this.mu.Lock()
	this.clients[svc.name] = svc
	this.mu.Unlock()

	if err := svc.start(); err != nil {
		svc.stop()
		return nil, err
	}

	glog.Infof("(%s, ip %v) server/handleConnection: Connection established.", svc.cid(), svc.ip)

	return svc, nil
}

func (this *Server) DeleteExistingSvc(c *client)  {

	this.mu.Lock()
	cid := c.name
	delete(this.clients, cid)
	this.mu.Unlock()
}
