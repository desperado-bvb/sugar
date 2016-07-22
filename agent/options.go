package agent

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/desperado-bvb/sugar/conf"
)


type Options struct {
	Host               string        `json:"addr"`
	KeepAlive 	   uint16 	 `json:"Keepalive"`
	Cluster        	   string        `json:"cluster_addr"`
	DataCenter         string        `json:"cluster_username"`
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
		case "host", "net":
			opts.Host = v.(string)
		case "keepalive":
			opts.KeepAlive = uint16(v.(int64))
		case "cluster":
			opts.Cluster = v.(string)
		case "datacenter":
			opts.DataCenter = v.(string)
		case "pidfile", "pid_file":
			opts.PidFile = v.(string)
		case "no_sigs":
                        opts.NoSigs = int(v.(int64))
		}
	}
	return opts, nil
}


func processOptions(opts *Options) {
	if opts.KeepAlive == 0 {
		opts.KeepAlive = DEFAULT_KEEPAlIVE
	}
}
