package login

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/redis/go-redis/v9"
	ini "gopkg.in/ini.v1"
)

var (
	cfg       *ini.File
	sec       *ini.Section
	err       error
	aesEnc    *utils.AesEncrypt
	aesStatus bool
	env       string
	apiUrl    string
	currency  string
	accessKey string
	proxyUrl  string
	headUrl   string

	cdnUrl          string
	rdb             *redis.Client
	RDB_KVSTORE_KEY = "KVSTORE:"
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

	//数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()
	data.InitLoginMgo(host, port, user, passwd, dbname)
	// redis init
	InitRedis()
	// nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	err = mq.InitNats(natsUrl)
	if err != nil {
		panic(err)
	}
	//初始化
	aesInit()
	config.Init2Game()
	// JtPayInit()
	// config.PayConfigInit(cfg)
	EfiInit(cfg)
	env = cfg.Section("env").Key("environment").Value()
	if env == "dev" {
		apiUrl = cfg.Section("ng").Key("devApiUrl").Value()
	} else {
		apiUrl = cfg.Section("ng").Key("procApiUrl").Value()
	}
	currency = cfg.Section("ng").Key("currency").Value()
	accessKey = cfg.Section("ng").Key("accessKey").Value()
	proxyUrl = cfg.Section("ng").Key("proxyUrl").Value()
	headUrl = cfg.Section("ng").Key("headUrl").Value()
	// cdn
	cdnUrl = cfg.Section("cdn").Key("url").Value()
	//启动服务
	bind := cfg.Section("login").Key("bind").Value()
	kind := cfg.Section("login").Key("kind").Value()
	NewRemote(bind, kind)
	//监听地址
	addr := cfg.Section("login").Key("addr").Value()
	//启动监听
	go Start(addr)
	go Wxmp()

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

func InitRedis() {
	redisAddr := cfg.Section("redis").Key("addr").Value()
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       1,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
}
