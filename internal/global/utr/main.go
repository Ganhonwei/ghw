package utr

import (
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"os"
	"os/signal"
	"runtime"
	"time"

	ini "gopkg.in/ini.v1"
)

var (
	cfg      *ini.File
	err      error
	env      string
	proxyUrl string
)

type Service struct{}

func (s *Service) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer glog.Flush()
	//日志定义
	glog.Init()

	//加载配置
	cfg, err = ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}

	cfg.BlockMode = false //只读

	//开发环境
	env = cfg.Section("env").Key("environment").Value()
	proxyUrl = cfg.Section("utr").Key("proxyUrl").Value()
	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}

	initNatsConsumer()

	glog.Infof("global utr service started")

	signalListen()

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
