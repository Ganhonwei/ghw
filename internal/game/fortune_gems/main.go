package fortune_gems

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data/mq"
	"goserver/pkg/flags"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myorderid"
	"goserver/pkg/table"

	jsoniter "github.com/json-iterator/go"
	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	node string

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

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	t := table.GetTables()
	_ = t

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

	//init orderid
	addr := cfg.Section("redis").Key("addr").Value()
	myorderid.InitOrderID(addr, 0)

	node = flags.Node
	//启动服务
	if node == "" {
		panic("unknown node")
	}
	env = cfg.Section("env").Key("environment").Value()
	nodeName = cfg.Section("game.fortune_gems." + node).Name()
	bind := cfg.Section(nodeName).Key("bind").Value()
	kind := cfg.Section(nodeName).Key("kind").Value()
	NewRemote(bind, kind)
	//初始化
	// config.Init2Game()

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
