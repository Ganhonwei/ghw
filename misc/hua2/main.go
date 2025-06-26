package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"time"

	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"

	jsoniter "github.com/json-iterator/go"
	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	node       = flag.String("node", "", "If non-empty, start with this node")
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile = flag.String("memprofile", "", "write mem profile to file")

	nodeName string
	env      string

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func main() {

	// u := data.User{}
	// u.Userid = "111"

	// intp := interp.New(interp.Options{}) // 初始化一个 yaegi 解释器
	// intp.Use(stdlib.Symbols)             // 允许脚本调用（几乎）所有的 Go 官方 package 代码
	// intp.Use(map[string]map[string]reflect.Value{
	// 	"goserver/pkg/data/data": {
	// 		"User": reflect.ValueOf((*data.User)(nil)),
	// 	},
	// })

	// intp.EvalPath("./script/hua.go") // src 就是上面的 Go 代码字符串

	// v, _ := intp.Eval("main.Fib")
	// fu := v.Interface().(func(int) int)

	// fmt.Println("Fib(35) =", fu(35))

	// v2, _ := intp.Eval("main.ChangeUser")
	// fu2 := v2.Interface().(func(*data.User))
	// fu2(&u)
	// fmt.Println(u)

	runtime.GOMAXPROCS(runtime.NumCPU())
	defer glog.Flush()
	//日志定义
	glog.Init()

	// zlog.Init("./zlogs", "hua2.log")
	// defer zlog.Sync()

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	t := table.GetTables()
	_ = t

	// name := t.TblTPRoom.Get(1001).Name
	// _ = name

	//性能监控
	pprofMonitor()
	//加载配置
	cfg, err = ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}
	cfg.BlockMode = false //只读

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	err = mq.InitNats(natsUrl)
	if err != nil {
		panic(err)
	}

	//启动服务
	if *node == "" {
		panic("unknown node")
	}
	env = cfg.Section("env").Key("environment").Value()
	nodeName = cfg.Section("game.hua_2" + *node).Name()
	bind := cfg.Section(nodeName).Key("bind").Value()
	kind := cfg.Section(nodeName).Key("kind").Value()
	NewRemote(bind, kind)
	//初始化
	config.Init2Game()

	handler.StartReadinessProbe(cfg, "20000")
	signalListen()
	//关闭服务
	Stop()
	//延迟等待
	<-time.After(10 * time.Second) //延迟关闭
}

func signalListen() {
	c := make(chan os.Signal)
	//signal.Notify(c)
	signal.Notify(c, os.Interrupt, os.Kill) //监听SIGINT和SIGKILL信号
	//signal.Stop(c)
	for {
		s := <-c
		glog.Error("get signal:", s)
		return
	}
}

func pprofMonitor() {
	// flag.Parse()
	glog.Notice("start pprof monitor")
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			glog.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()

		if err := pprof.StartCPUProfile(f); err != nil {
			glog.Fatal("could not start CPU profile: ", err)
		}
		glog.Notice("start CPU profile: ", utils.LocalTime())
		defer pprof.StopCPUProfile()
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			glog.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()

		if err := pprof.WriteHeapProfile(f); err != nil {
			glog.Fatal("cound not write memory profile: ", err)
		}
		glog.Notice("start MEM profile: ", utils.LocalTime())
	}
}
