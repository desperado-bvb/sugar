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
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//fmt.Println(runtime.NumCPU())

	opt, err := master.ProcessConfigFile(cfgfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("read configuration success")

	svr := master.New(opt)
	svr.Start()
}
