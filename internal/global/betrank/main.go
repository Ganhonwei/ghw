package betrank

import (
	"context"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/ini.v1"
)

// 打码排行榜
var (
	cfg *ini.File

	rdb     *redis.Client
	rdbHead *redis.Client

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
	var err error
	cfg, err = ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}

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

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	if err = initNatsRouter(); err != nil {
		panic(err)
	}

	// init nats producer
	ck.InitNatsProducer()

	redisAddr := cfg.Section("redis").Key("addr").Value()
	if rdb, err = InitRedis(redisAddr, 0); err != nil {
		panic(err)
	}
	if rdbHead, err = InitRedis(redisAddr, 3); err != nil {
		panic(err)
	}

	//数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()
	data.InitMgo(host, port, user, passwd, dbname)

	// init clickhouse
	// ck.InitClickhouse(ck.GetConfigFromIni(cfg))
	// ck.InitChannelProducerConsumer()

	InitBetRank()

	bind := cfg.Section("activity.betrank").Key("bind").Value()
	kind := cfg.Section("activity.betrank").Key("kind").Value()
	NewRemote(bind, kind)

	handler.StartReadinessProbe(cfg, "20000")
	handler.StartPrestopCallback(cfg, "20001", Stop)
	signalListen() //监听关闭信号
	//关闭服务
	// Stop()
	//延迟等待
	<-time.After(10 * time.Second) //延迟关闭
}

func signalListen() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP) //监听SIGINT和SIGKILL信号
	s := <-sigChan
	glog.Error("get signal:", s)
}

func InitRedis(redisAddr string, db int) (rdb *redis.Client, err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err = rdb.Ping(ctx).Result()
	return
}
