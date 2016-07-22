package main

import (
	"fmt"
	"flag"
	"github.com/desperado-bvb/sugar/agent"
)

const (
	xcfgfile string = "opt1.cfg"
)

func init() {
	flag.Parse()
}

func main() {

	opt, err := agent.ProcessConfigFile(xcfgfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	svr := agent.New(opt)
	svr.Start()
}
