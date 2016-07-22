sugar
====

非常非常非常非常简单版本的进程管理

## Sugar Master 
负责管理所有的agent上面的进程启动和停止，（todo: 备份，NameService，load rebalance, 可执行文件管理）

```go
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
```
配置文件
```cfg
host:"localhost"            //监听agent连接host
port:10000                  //监听agent连接port
http:"localhost:10001"      //API接口
pidfile:"pid.file"          
no_sigs:1                  //是否捕捉信号
```


## sugar agent
单机进程管理守护进程，负责执行master进程指令（to: 监听进程异常退出上报，重启；上报自身节点资源使用情况；重启后故障恢复等）

```go
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

	opt, err := agent.ProcessConfigFile(cfgfile)
	if err != nil {
		fmt.Println(err)
		return
	}

	svr := agent.New(opt)
	svr.Start()
}
```
配置文件
```cfg
host:"localhost:10000"         //master端口
cluster:"BigData"              //Cluster
datacenter:"asian"             //数据中心 
pidfile:"pid1.file"
no_sigs:1
```
## sugar 已实现功能以及使用查询样例
```cfg
1 agent注册   
2 node查询              curl "http://localhost:10001/servers"
3 node的进程查询         curl "http://localhost:10001/query?name=cluster/datacenter/agentid"
4 node的进程启动         curl "http://localhost:10001/start?name=cluster/datacenter/agentid&cmd=cmd"
5 node的进程关闭         curl "http://localhost:10001/stop?name=cluster/datacenter/agentid&pid=pid"
```



