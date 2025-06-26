package logger

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"goserver/pkg/table"

	jsoniter "github.com/json-iterator/go"
	"github.com/oschwald/geoip2-golang"
	ini "gopkg.in/ini.v1"
)

var (
	cfg          *ini.File
	sec          *ini.Section
	err          error
	locationName string
	location     *time.Location

	env  string
	json = jsoniter.ConfigCompatibleWithStandardLibrary

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

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	// 加载时区
	locationName = cfg.Section("env").Key("location").Value()
	location, err = time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}

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

	//clickhouse kafka
	// ck.InitClickhouse(ck.GetConfigFromIni(cfg))
	// ck.InitKafkaProducer(cfg)
	// ck.InitKafkaConsumer(cfg)
	// ck.InitChannelProducerConsumer()

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	// init nats producer
	ck.InitNatsProducer()

	// init redis
	addr := cfg.Section("redis").Key("addr").Value()
	myredis.InitRedis(addr, 0)

	//初始化ip库
	initIPdat()
	//启动服务
	bind := cfg.Section("logger").Key("bind").Value()
	kind := cfg.Section("logger").Key("kind").Value()
	NewRemote(bind, kind)

	handler.StartReadinessProbe(cfg, "20000")
	handler.StartPrestopCallback(cfg, "20001", Stop)
	signalListen()
	//关闭服务
	//Stop()

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
