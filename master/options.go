package master

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/desperado-bvb/sugar/conf"
)


type Options struct {
	Host               string        `json:"addr"`
	Port               int           `json:"port"`
	HTTPAddress        string        `json:"http"`
	KeepAlive 	   int 	 	 `json:"Keepalive"`
	ConnectTimeout 	   int  	 `json:"connect_timeout"`
	AckTimeout         int 		 `json:"ack_timeout"`
	TimeoutRetries 	   int 		 `json:"timeout_retries"`
	MaxConn            int           `json:"max_connections"`
	NoSigs             int           `json:"no_sigs"`
	PidFile            string        `json:"-"`
}

func ProcessConfigFile(configFile string) (*Options, error) {
	opts := &Options{}

	if configFile == "" {
		return opts, nil
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}

	m, err := conf.Parse(string(data))
	if err != nil {
		return nil, err
	}

	for k, v := range m {
		switch strings.ToLower(k) {
		case "port":
			opts.Port = int(v.(int64))
		case "host", "net":
			opts.Host = v.(string)
		case "http":
			opts.HTTPAddress = v.(string)
		case "keepalive":
			opts.KeepAlive = int(v.(int64))
		case "connect_timeout":
			opts.ConnectTimeout = int(v.(int64))
		case "ack_timeout":
			opts.AckTimeout = int(v.(int64))
		case "timeout_retries":
			opts.TimeoutRetries = int(v.(int64))
		case "max_connections", "max_conn":
			opts.MaxConn = int(v.(int64))
		case "pidfile", "pid_file":
			opts.PidFile = v.(string)
		case "no_sigs":
			opts.NoSigs = int(v.(int64))
		}
	}
	return opts, nil
}

func processOptions(opts *Options) {
	if opts.Host == "" {
		opts.Host = "0.0.0.0"
	}

	if opts.HTTPAddress == "" {
		opts.HTTPAddress = "0.0.0.0:10001"
	}

	if opts.Port == 0 {
		opts.Port = DEFAULT_PORT
	}

	if opts.MaxConn == 0 {
		opts.MaxConn = DEFAULT_MAX_CONNECTIONS
	}

	if opts.KeepAlive == 0 {
		opts.KeepAlive = DEFAULT_KEEPAlIVE
	}

	if opts.ConnectTimeout == 0 {
		opts.ConnectTimeout = DEFAULT_CONNECT_TIMEOUT
	}

	if opts.AckTimeout == 0 {
		opts.AckTimeout = DEFAULT_ACKT_IMEOUT
	}

	if opts.TimeoutRetries == 0 {
		opts.TimeoutRetries = DEFAULT_TIMEOUT_RETRIES
	}

}
