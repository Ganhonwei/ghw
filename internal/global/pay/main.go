package pay

import (
	"goserver/internal/global/pay/config"
	"goserver/internal/global/pay/service"
	tasks "goserver/internal/global/pay/task"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	aesEnc *utils.AesEncrypt

	aesStatus bool

	redisAddr string
	Country   string
)

type Service struct {
}

func (s *Service) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer glog.Flush()
	//日志定义
	glog.Init()
	//加载配置
	cfg, err := ini.Load("config/app.conf")
	if err != nil {
		panic(err)
	}
	// cfg.BlockMode = false //只读
	//数据库连接
	service.InitMgo(cfg)
	//初始化
	service.InitWeb(cfg)
	// aesInit()

	// load nsq config
	conf, err := ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}
	service.InitNsq(conf)
	service.InitPayChannel()

	// init nats
	natsUrl := conf.Section("nats").Key("url").Value()
	err = mq.InitNats(natsUrl)
	if err != nil {
		panic(err)
	}

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}

	service.Env = cfg.Section("env").Key("environment").Value()
	Country = cfg.Section("env").Key("country").Value()
	//初始化ayncq
	redisAddr = cfg.Section("").Key("redis.addr").String()
	if redisAddr == "" {
		panic("init ayncq redis fail")
	}
	go InitAsynqServer(redisAddr)
	err1 := tasks.InitAsynqClient(redisAddr)
	if err1 != nil {
		panic(err1)
	}

	config.PayConfigInit(cfg)
	// env = cfg.Section("env").Key("environment").Value()

	//监听地址
	addr := cfg.Section("").Key("pay.addr").String()
	bind := cfg.Section("").Key("pay.bind").String()
	kind := cfg.Section("").Key("pay.kind").String()
	NewRemote(bind, kind)
	//启动监听
	go Start(addr)

	handler.StartReadinessProbe(cfg, "20000")
	handler.StartPrestopCallback(cfg, "20001", Stop)
	signalListen() //监听关闭信号
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

// 加密初始化
func aesInit() {
	aesEnc = new(utils.AesEncrypt)
	key := cfg.Section("login").Key("key").Value()
	aesEnc.SetKey([]byte(key))
	aesStatus = cfg.Section("login").Key("status").MustBool(false)
}

// 加密
func aesEn(doc string) (arrEncrypt []byte) {
	arrEncrypt, err = aesEnc.Encrypt([]byte(doc))
	if err != nil {
		glog.Errorf("arrEncrypt: %s", doc)
	}
	return
}

// 解密
func aesDe(arrEncrypt []byte) (strMsg string) {
	bMsg, err := aesEnc.Decrypt(arrEncrypt)
	if err != nil {
		glog.Errorf("arrEncrypt: %s", string(arrEncrypt))
	}
	strMsg = string(bMsg)
	return
}

// 初始化ayncqserver
func InitAsynqServer(redisAddr string) {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			// Specify how many concurrent workers to use
			Concurrency: 10,
			// Optionally specify multiple queues with different priority.
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			// See the godoc for other configuration options
		},
	)
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.HandleFunc(tasks.TypeWithdrawBadDelivery, tasks.HandleBadWithdrawDeliveryTask)
	mux.HandleFunc(tasks.TypePayLogDelivery, tasks.HandlePayLogDeliveryTask)
	// ...register other handlers...

	if err := srv.Run(mux); err != nil {
		glog.Fatalf("could not run server: %v", err)
	}
}
