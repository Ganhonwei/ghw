package gate

import (
	"goserver/pkg/data/ck"
	"goserver/pkg/data/mq"
	"goserver/pkg/flags"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"goserver/pkg/data"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"

	jsoniter "github.com/json-iterator/go"
	"github.com/oschwald/geoip2-golang"
	ini "gopkg.in/ini.v1"
)

var (
	cfg          *ini.File
	sec          *ini.Section
	ipClient     *geoip2.Reader
	err          error
	locationName string
	location     *time.Location

	aesEnc *utils.AesEncrypt

	aesStatus bool

	node string

	nodeName string

	env     string
	country string

	json = jsoniter.ConfigCompatibleWithStandardLibrary

	//HeadImagList 默认头像
	HeadImagList []data.RegistPhoto

	payurl      string
	withdrawurl string

	// 大富翁中奖牌
	scratchTick data.ScratchJackpot
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

	// 加载时区
	locationName = cfg.Section("env").Key("location").Value()
	location, err = time.LoadLocation(locationName)
	if err != nil {
		panic(err)
	}
	// 地区标识
	country = cfg.Section("env").Key("country").Value()

	cfg.BlockMode = false //只读
	//初始化
	aesInit()
	HeadImagList = data.RegistPhotos()
	//nsq初始化
	InitNsq()

	// init nats
	natsUrl := cfg.Section("nats").Key("url").Value()
	err = mq.InitNats(natsUrl)
	if err != nil {
		panic(err)
	}
	ck.InitClickhouse(ck.GetConfigFromIni(cfg))

	//加载配置表
	err = table.LoadTables()
	if err != nil {
		panic(err)
	}
	node = flags.Node
	//启动服务
	if node == "" {
		panic("unknown node")
	}
	env = cfg.Section("env").Key("environment").Value()
	nodeName = cfg.Section("gate.node" + node).Name()
	bind := cfg.Section(nodeName).Key("bind").Value()
	kind := cfg.Section(nodeName).Key("kind").Value()
	NewRemote(bind, kind)
	//配置初始化
	appid := cfg.Section("weixin").Key("appid").Value()
	appsecret := cfg.Section("weixin").Key("appsecret").Value()
	appkey := cfg.Section("weixin").Key("appkey").Value()
	mchid := cfg.Section("weixin").Key("mchid").Value()
	pattern := cfg.Section("weixin").Key("notifyPattern").Value()
	notifyURL := cfg.Section("weixin").Key("notifyUrl").Value()
	// 充值提现
	payurl = cfg.Section("recharge").Key("payurl").Value()
	withdrawurl = cfg.Section("recharge").Key("withdrawurl").Value()

	config.Init2Gate(appid, appsecret, appkey, mchid, pattern, notifyURL)
	//充值配置初始化
	// config.PayConfigInit(cfg)
	//初始化ip库
	initIPdat()
	//变量初始化
	secret := cfg.Section("gate").Key("secret").Value()
	login.TokenInit(secret)
	// 日志库
	host := cfg.Section("mongod.log").Key("host").Value()
	port := cfg.Section("mongod.log").Key("port").Value()
	user := cfg.Section("mongod.log").Key("user").Value()
	passwd := cfg.Section("mongod.log").Key("passwd").Value()
	dbname := cfg.Section("mongod.log").Key("name").Value()
	data.InitLogMgo(host, port, user, passwd, dbname)
	//wsServer
	addr := cfg.Section(nodeName).Key("addr").Value()
	addr2 := cfg.Section(nodeName).Key("addr2").Value()
	wsServer := new(WSServer)
	wsServer.Addr = addr
	wsServer.Addr2 = addr2
	if wsServer != nil {
		wsServer.Start()
	}

	callback := func() {
		if wsServer != nil {
			wsServer.Close()
		}
		defer ipClient.Close()

		Stop()
	}

	handler.StartReadinessProbe(cfg, "20000")
	handler.StartPrestopCallback(cfg, "20001", callback)
	signalListen() //监听关闭信号
	//关闭服务
	//关闭websocket连接, 先关监听
	// if wsServer != nil {
	// 	wsServer.Close()
	// }
	// defer ipClient.Close()
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
	key := cfg.Section("gate").Key("key").Value()
	aesEnc.SetKey([]byte(key))
	aesStatus = cfg.Section("login").Key("status").MustBool(false)
}

// 加密
func aesEn(doc []byte) (arrEncrypt []byte) {
	arrEncrypt, err = aesEnc.Encrypt(doc)
	if err != nil {
		glog.Errorf("arrEncrypt: %s", string(doc))
	}
	return
}

// 解密
func aesDe(arrEncrypt []byte) (bMsg []byte) {
	bMsg, err = aesEnc.Decrypt(arrEncrypt)
	if err != nil {
		glog.Errorf("arrEncrypt: %s", string(arrEncrypt))
	}
	return
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

func isChina(d string) bool {
	return d == "China"
}

// ip拦截
func ipInterceptor(ip string, status int32) bool {
	if !config.SettingIsOpen(4, data.IPLIMIT) {
		return false
	}
	if config.GetIpWhiteByKey(ip) {
		// 在白名单中
		glog.Infof("ip in white lsit,ip:%s", ip)
		return false
	}
	if ipClient == nil {
		return true
	}

	addr := net.ParseIP(ip)
	record, err := ipClient.Country(addr)
	if err != nil {
		glog.Errorf("ip is err:%s", err)
		return false
	}
	if isChina(record.Country.Names["en"]) && status != 4 {
		return true
	}
	return false
}
