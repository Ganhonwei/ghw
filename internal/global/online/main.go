package online

import (
	"context"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/redis/go-redis/v9"
	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	err error

	rdb *redis.Client
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

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	if err = mq.InitNats(natsUrl); err != nil {
		panic(err)
	}
	if err = initNatsRouter(); err != nil {
		panic(err)
	}
	initNatsConsumer()

	//数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()
	data.InitMgo(host, port, user, passwd, dbname)

	redisAddr := cfg.Section("redis").Key("addr").Value()
	if rdb, err = InitRedis(redisAddr, 0); err != nil {
		panic(err)
	}

	StartActiveUserChecker()

	glog.Infof("global alert service started")

	signalListen()

	//延迟等待
	<-time.After(10 * time.Second) //延迟关闭
}

func signalListen() {
	c := make(chan os.Signal, 1)
	//signal.Notify(c)
	signal.Notify(c, os.Interrupt) //监听SIGINT和SIGKILL信号
	//signal.Stop(c)
	for {
		s := <-c
		glog.Error("get signal:", s)
		return
	}
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
