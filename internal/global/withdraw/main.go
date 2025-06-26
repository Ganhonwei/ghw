package withdraw

import (
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"goserver/pkg/table"
	"os"
	"os/signal"
	"runtime"
	"time"

	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	err error
	env string

	locationName string
	location     *time.Location
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

	// 加载时区
	locationName = cfg.Section("env").Key("location").Value()
	location, err = time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	glog.Infof("load time location: %s", locationName)

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	//开发环境
	env = cfg.Section("env").Key("environment").Value()
	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}

	// redis
	addr := cfg.Section("redis").Key("addr").Value()
	myredis.InitRedis(addr, 0)

	//数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()
	data.InitMgo(host, port, user, passwd, dbname)

	initNatsConsumer()

	start()
	glog.Infof("global utr service started")

	signalListen()
	stop()

	//延迟等待
	<-time.After(10 * time.Second) //延迟关闭
}

func signalListen() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt) //监听SIGINT和SIGKILL信号
	for {
		s := <-c
		glog.Error("get signal:", s)
		return
	}
}

func stop() {
	ctx := c.Stop()
	<-ctx.Done()
}
