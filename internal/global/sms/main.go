package sms

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	ini "gopkg.in/ini.v1"
)

var (
	cfg *ini.File
	sec *ini.Section
	err error

	aesEnc *utils.AesEncrypt

	aesStatus bool

	env string
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
	host := cfg.Section("mongod.log").Key("host").Value()
	port := cfg.Section("mongod.log").Key("port").Value()
	user := cfg.Section("mongod.log").Key("user").Value()
	passwd := cfg.Section("mongod.log").Key("passwd").Value()
	dbname := cfg.Section("mongod.log").Key("name").Value()
	data.InitSmsMgo(host, port, user, passwd, dbname)
	//初始化
	aesInit()
	//短信初始化
	SmsInit()
	env = cfg.Section("env").Key("environment").Value()
	//启动服务
	bind := cfg.Section("sms").Key("bind").Value()
	kind := cfg.Section("sms").Key("kind").Value()
	NewRemote(bind, kind)
	//监听地址
	addr := cfg.Section("sms").Key("addr").Value()
	//启动监听
	go Start(addr)
	signalListen() //监听关闭信号
	//关闭服务
	Stop()
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
	key := cfg.Section("sms").Key("key").Value()
	aesEnc.SetKey([]byte(key))
	aesStatus = cfg.Section("sms").Key("status").MustBool(false)
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
