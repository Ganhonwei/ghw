package main

import (
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"os"
	"os/signal"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/ini.v1"
)

// 临时统计导出为excel

var (
	cfg      *ini.File
	sec      *ini.Section
	err      error
	location *time.Location

	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ExportDir = "./temp_stats_export"
)

func main() {
	location, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}

	glog.Info("temp-stats start")
	// 创建文件导出目录
	if !utils.FileExists(ExportDir) {
		if e1 := os.Mkdir(ExportDir, os.ModePerm); e1 != nil {
			glog.Error("mkdir export dir error: ", e1)
		}
	}

	// 执行统计
	// runStat("5-12CRASH超过10局玩家统计", HandleExportCrash50)

	// runStat("各类用户游戏局数统计", StatPlayerRounds)

	// runStat("各类用户游戏局数范围统计", RoundStats)

	// runStat("全量用户游戏类型局数统计", UserGameDataFull)

	// runStat("回合数201统计", RoundStats201)

	// runStat("4月份注册玩家统计", StatPlayersM4)

	// runStat("少于5局用户统计", StatLess5Rounds)

	// runStat("rummy游戏点数统计", StatsRummyScore)
	runStat("注册用户前10局的游戏分布输赢", Stat7_1to7_14Round10)

	time.Sleep(time.Second * 3)
}

func runStat(name string, f func()) {
	glog.Infof("task exec start %s", name)
	sec := time.Now().Unix()
	f()
	glog.Infof("task exec finish  %s use %ds", name, time.Now().Unix()-sec)
}

func init() {
	// init config, init mongo
	defer glog.Flush()
	//日志定义
	glog.Init()

	//加载配置
	cfg, err = ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}

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
}

func signalListen() {
	c := make(chan os.Signal)
	//signal.Notify(c)
	signal.Notify(c, os.Interrupt, os.Kill) //监听SIGINT和SIGKILL信号
	//signal.Stop(c)
	for {
		s := <-c
		glog.Error("get signal:", s)
		return
	}
}
