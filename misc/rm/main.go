package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"time"

	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"

	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	node = flag.String("node", "", "If non-empty, start with this node")

	nodeName string
	env      string

	json        = jsoniter.ConfigCompatibleWithStandardLibrary
	redisClient *redis.Client
)

func main() {
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
	//启动服务
	if *node == "" {
		panic("unknown node")
	}
	env = cfg.Section("env").Key("environment").Value()
	nodeName = cfg.Section("game.rummy" + *node).Name()
	bind := cfg.Section(nodeName).Key("bind").Value()
	kind := cfg.Section(nodeName).Key("kind").Value()

	NewRemote(bind, kind)
	//初始化
	config.Init2Game()

	redisAddr := cfg.Section("redis").Key("addr").Value()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       2,
	})

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	err = mq.InitNats(natsUrl)
	if err != nil {
		panic(err)
	}

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
