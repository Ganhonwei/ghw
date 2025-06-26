package hua

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data/mq"
	"goserver/pkg/flags"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"

	jsoniter "github.com/json-iterator/go"
	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	node string
	// cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	// memprofile = flag.String("memprofile", "", "write mem profile to file")

	nodeName string
	env      string

	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type Service struct{}

func (s *Service) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer glog.Flush()
	//日志定义
	glog.Init()

	// zlog.Init("./zlogs", "hua.log")
	// defer zlog.Sync()

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	t := table.GetTables()
	_ = t

	//性能监控
	// pprofMonitor()
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

	node = flags.Node
	//启动服务
	if node == "" {
		panic("unknown node")
	}
	env = cfg.Section("env").Key("environment").Value()
	nodeName = cfg.Section("game.hua" + node).Name()
	bind := cfg.Section(nodeName).Key("bind").Value()
	kind := cfg.Section(nodeName).Key("kind").Value()
	NewRemote(bind, kind)
	//初始化
	config.Init2Game()

	handler.StartReadinessProbe(cfg, "20000")
	handler.StartPrestopCallback(cfg, "20001", Stop)
	signalListen()
	//关闭服务
	// Stop()
	//延迟等待
	<-time.After(10 * time.Second) //延迟关闭
}

func signalListen() {
	c := make(chan os.Signal, 1)
	//signal.Notify(c)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM) //监听SIGINT和SIGKILL信号
	//signal.Stop(c)
	for {
		s := <-c
		glog.Error("get signal:", s)
		return
	}
}

// func pprofMonitor() {
// 	// flag.Parse()
// 	glog.Notice("start pprof monitor")
// 	if *cpuprofile != "" {
// 		f, err := os.Create(*cpuprofile)
// 		if err != nil {
// 			glog.Fatal("could not create CPU profile: ", err)
// 		}
// 		defer f.Close()

// 		if err := pprof.StartCPUProfile(f); err != nil {
// 			glog.Fatal("could not start CPU profile: ", err)
// 		}
// 		glog.Notice("start CPU profile: ", utils.LocalTime())
// 		defer pprof.StopCPUProfile()
// 	}

// 	if *memprofile != "" {
// 		f, err := os.Create(*memprofile)
// 		if err != nil {
// 			glog.Fatal("could not create memory profile: ", err)
// 		}
// 		defer f.Close()

// 		if err := pprof.WriteHeapProfile(f); err != nil {
// 			glog.Fatal("cound not write memory profile: ", err)
// 		}
// 		glog.Notice("start MEM profile: ", utils.LocalTime())
// 	}
// }
