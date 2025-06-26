package dbms

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
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

	env     string
	country string
	json    = jsoniter.ConfigCompatibleWithStandardLibrary

	payurl      string
	withdrawurl string
	ipClient    *geoip2.Reader
	smsService  string
)

type Service struct{}

func (s *Service) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer glog.Flush()
	//日志定义
	glog.Init()
	// zlog.Init("./zlogs", "dbms.log")
	// defer zlog.Sync()

	//加载配置
	cfg, err = ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}

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

	// 地区标识
	country = cfg.Section("env").Key("country").Value()

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
	//clickhouse kafka
	ck.InitClickhouse(ck.GetConfigFromIni(cfg))
	// ck.InitKafkaProducer(cfg)
	// ck.InitKafkaConsumer(cfg)
	// ck.InitChannelProducerConsumer()

	//充值提现
	payurl = cfg.Section("recharge").Key("payurl").Value()
	withdrawurl = cfg.Section("recharge").Key("withdrawurl").Value()

	//sms服务
	smsService = cfg.Section("sms-service").Key("addr").Value()

	data.InitLogMgo(host, port, user, passwd, dbname)

	//游戏配置数据初始化
	handler.InitGame()

	//配置初始化
	config.ConfigInit()
	// config.PayConfigInit(cfg)

	//库存数据初始化
	handler.InitStock()
	//nsq初始化
	// InitNsq()

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	// init nats producer
	ck.InitNatsProducer()

	//初始化ip库
	initIPdat()

	//init redis
	addr := cfg.Section("redis").Key("addr").Value()
	myredis.InitRedis(addr, 0)

	//启动服务
	bind := cfg.Section("dbms").Key("bind").Value()
	kind := cfg.Section("dbms").Key("kind").Value()
	room := cfg.Section("dbms").Key("room").Value()
	role := cfg.Section("dbms").Key("role").Value()
	// logger := cfg.Section("dbms").Key("logger").Value()
	report := cfg.Section("dbms").Key("report").Value()
	NewRemote(bind, kind, room, role, report)
	//test
	// Test()
	// FBTest()

	handler.StartReadinessProbe(cfg, "20000")
	handler.StartPrestopCallback(cfg, "20001", Stop)
	signalListen()
	//关闭服务
	// Stop()
	//关闭nsq
	// StopNsq()
	// data.Close()     //数据库断开
	// ipClient.Close() //ip库断开
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
