package main

import (
	"fmt"
	"flag"
	"github.com/desperado-bvb/sugar/master"
)

const (
	cfgfile string = "opt.cfg"
)

func init() {
	flag.Parse()
}

func main() {
	opt, err := master.ProcessConfigFile(cfgfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	svr := master.New(opt)
	svr.Start()
}
