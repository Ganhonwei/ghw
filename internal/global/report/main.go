package report

import (
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/oschwald/geoip2-golang"
	ini "gopkg.in/ini.v1"
)

var (
	cfg      *ini.File
	err      error
	env      string
	ipClient *geoip2.Reader
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

	//数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()
	data.InitMgo(host, port, user, passwd, dbname)

	//日志库
	host = cfg.Section("mongod.log").Key("host").Value()
	port = cfg.Section("mongod.log").Key("port").Value()
	user = cfg.Section("mongod.log").Key("user").Value()
	passwd = cfg.Section("mongod.log").Key("passwd").Value()
	dbname = cfg.Section("mongod.log").Key("name").Value()
	data.InitLogMgo(host, port, user, passwd, dbname)

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	ck.InitNatsProducer()

	//初始化ip库
	initIPdat()

	//init redis
	addr := cfg.Section("redis").Key("addr").Value()
	myredis.InitRedis(addr, 0)

	initNatsConsumer()

	glog.Infof("global report service started")

	signalListen()

	data.Close()     //数据库断开
	ipClient.Close() //ip库断开

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

// 初始化ip库
func initIPdat() {
	mmdbBytes, err := os.ReadFile("config/GeoLite2-City.mmdb")
	if err != nil {
		panic(err)
	}
	ipClient, err = geoip2.FromBytes(mmdbBytes)
	if err != nil {
		panic(err)
	}
}
