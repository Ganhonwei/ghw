package main

import (
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/game/config"
	"goserver/pkg/glog"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/globalsign/mgo"
	"gopkg.in/ini.v1"
)

var (
	location *time.Location
)

// mongo 同步 clickhouse
func main() {
	// defer glog.Flush()
	//日志定义
	glog.Init()

	//加载配置
	cfg, err := ini.Load("config/conf.ini")
	if err != nil {
		panic(err)
	}

	location, err = time.LoadLocation("Asia/Kolkata")
	if err != nil {
		panic(err)
	}

	initLogMongo(cfg)
	initDataMongo(cfg)

	ck.InitClickhouse(ck.GetConfigFromIni(cfg))
	// ck.AutoMigrateTables()

	wg := &sync.WaitGroup{}
	// go SyncFullUser(wg)
	// go SyncFullTradeRecord(wg)
	go SyncFullWithdrawRecord(wg)
	// go SyncFullLogLogin(wg)
	// go SyncFullLogWater(wg)
	// go SyncFullDetail(wg)
	// go SyncFullExternalBet(wg)
	// go SyncFullExternalRewards(wg)
	// go SyncFullExternalCancel(wg)
	// go SyncFullLogVbDiamond(wg)
	// go SyncFullLogOutDiamond(wg)
	// go SyncFullActivityBetRankPrize(wg)
	// go SyncFullActivityTurnDrawLog(wg)
	// go SyncFullActivityTurnPrizeLog(wg)
	// go SyncFullShareAgentIncomeRecord(wg)
	// go SyncFullShareAgentIncomeRecordTackLog(wg)
	// go SyncFullVolatilitySubsidy(wg)
	// go SyncFullOnlineUsersLog(wg)
	// go SyncFullUserCoupon(wg)

	time.Sleep(time.Second)
	wg.Wait()
}

func initLogMongo(cfg *ini.File) {
	//日志库
	host := cfg.Section("mongod.log").Key("host").Value()
	port := cfg.Section("mongod.log").Key("port").Value()
	user := cfg.Section("mongod.log").Key("user").Value()
	passwd := cfg.Section("mongod.log").Key("passwd").Value()
	dbname := cfg.Section("mongod.log").Key("name").Value()
	data.InitLogMgo(host, port, user, passwd, dbname)
}

func initDataMongo(cfg *ini.File) {
	// //数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()

	data.InitMgo(host, port, user, passwd, dbname)
	// 初始化游戏房间
	config.InitGame()
}

func initSSHDataMongo(cfg *ini.File) {
	// //数据库连接
	host := cfg.Section("mongod").Key("host").Value()
	port := cfg.Section("mongod").Key("port").Value()
	user := cfg.Section("mongod").Key("user").Value()
	passwd := cfg.Section("mongod").Key("passwd").Value()
	dbname := cfg.Section("mongod").Key("name").Value()
	sshON, _ := cfg.Section("mongod").Key("sshON").Bool()
	sshKey := cfg.Section("mongod").Key("sshKey").Value()
	sshUser := cfg.Section("mongod").Key("sshUser").Value()
	sshAddr := cfg.Section("mongod").Key("sshAddr").Value()

	fmt.Println("mongo: ", host, port, user, passwd)

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{fmt.Sprintf("%s:%s", host, port)},
		Database: dbname,
		Username: user,
		Password: passwd,
	}
	if user != "" && passwd != "" {
		dialInfo.Source = "admin"
		dialInfo.Mechanism = "SCRAM-SHA-1"
	}
	// mongo ssh
	if sshON {
		// SSH 配置
		k, err := os.ReadFile(sshKey)
		if err != nil {
			panic(err)
		}
		signer, err := ssh.ParsePrivateKey(k)
		if err != nil {
			panic(err)
		}

		sshConfig := &ssh.ClientConfig{
			User: sshUser,
			Auth: []ssh.AuthMethod{
				// ssh.Password("ssh_password"), // SSH 密码认证
				ssh.PublicKeysCallback(func() ([]ssh.Signer, error) { // SSH 私钥认证
					return []ssh.Signer{signer}, nil
				}),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 忽略主机密钥检查（仅供开发使用）
			Timeout:         10 * time.Second,            // SSH 连接超时
		}

		// 连接 SSH 服务器
		sshClient, err := ssh.Dial("tcp", sshAddr, sshConfig)
		if err != nil {
			panic(err)
		}
		// 创建本地监听器 (本地端口)
		localListener, err := sshClient.Dial("tcp", fmt.Sprintf("%s:%s", host, port)) // 远程 ClickHouse 地址及端口
		if err != nil {
			panic(err)
		}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return localListener, nil
		}
	}
	var err error
	data.Session, err = mgo.DialWithInfo(dialInfo)
	if err != nil {
		panic(err)
	}

	data.InitMgo(host, port, user, passwd, dbname)
	// 初始化游戏房间
	config.InitGame()
}
