package main

import (
	"fmt"
	"flag"
	"github.com/desperado-bvb/sugar/agent"
)

const (
	cfgfile string = "opt1.cfg"
)

func init() {
	flag.Parse()
}

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())
	//fmt.Println(runtime.NumCPU())

	opt, err := agent.ProcessConfigFile(cfgfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println("read configuration success")

	svr := agent.New(opt)
	svr.Start()
}
