package sync

import (
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	err error
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

	ck.InitClickhouse(ck.GetConfigFromIni(cfg))
	// 对齐ck表结构
	if err = ck.AutoMigrateTables(); err != nil {
		panic(err)
	}

	// 初始化游戏库
	initDataMongo(cfg)

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}

	// init nats consumer
	ck.InitNatsConsumer()

	glog.Info("global sync service started")

	signalListen()

	//延迟等待
	<-time.After(10 * time.Second) //延迟关闭
}

func initDataMongo(cfg *ini.File) {
	// //数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()

	data.InitMgo(host, port, user, passwd, dbname)
	// 初始化游戏房间
	config.InitGame()
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
