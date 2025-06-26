package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/gorilla/websocket"
)

type RobotServer struct {
	PendingWriteNum int    //等待写入消息长度
	MaxMsgLen       uint32 //最大消息长度

	conns      WebsocketConnSet //连接集合
	mutexConns sync.Mutex       //互斥锁
	wg         sync.WaitGroup   //同步机制

	//channel chan *pb.RobotMsg //消息通道
	channel chan interface{} //消息通道
	closeCh chan struct{}    // 关闭通道

	Name   string //注册节点名字
	phone1 string //注册登录账号
	phone2 string //模拟人机注册登录账号

	online  map[string]bool  //map[phone]状态,true=在线,
	offline map[string]bool  //map[phone]状态,true=离线,false=登录中
	unused  map[string]int64 //map[phone]chip 筹码不足的

	robotPool      map[int32][]*Robot //gtype -> 机器人池
	robotPoolMutex sync.Mutex         //机器人池锁

	simrobotPool      []*Robot   //模拟人机池
	simrobotPoolMutex sync.Mutex //模拟人机池

	robotId    int //机器人ID
	simrobotId int //模拟人机ID

	maxId int //最大id
	count int //每次登录

	robotNum    map[int32]int //人机数
	simrobotNum int           //模拟人机数

	loginFinish bool //登录是否完成

	rooms  map[string]int     //房间人数
	ltypes map[int32][]string //map[ltype][]{roomid..}

	hua    map[string]int
	lhd    int
	updown int
	crash  int

	msgCh  chan interface{} //消息通道
	stopCh chan struct{}    // 关闭通道
}

func (server *RobotServer) Start() {
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		glog.Infof("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 1024
		glog.Infof("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}

	server.conns = make(WebsocketConnSet)

	//初始化
	server.online = make(map[string]bool)
	server.offline = make(map[string]bool)
	server.unused = make(map[string]int64)
	server.rooms = make(map[string]int)
	server.ltypes = make(map[int32][]string)
	server.msgCh = make(chan interface{}, 100)
	server.stopCh = make(chan struct{})
	server.robotPool = make(map[int32][]*Robot)

	total := 200
	server.hua = map[string]int{
		"1001": total,
		"1002": total,
		"1003": total,
		"1004": total,
		"1005": total,
	}
	server.lhd = 400
	server.updown = 400
	server.crash = 200

	server.robotNum = map[int32]int{
		int32(pb.HUA):   0,
		int32(pb.JOKER): 0,
		int32(pb.AK47):  0,
		int32(pb.LHD):   0,
		int32(pb.SEVEN): 0,
		int32(pb.CRASH): 0,
		int32(pb.RUMMY): 0,
	}

	// server.robotNum = 2000
	if env == "dev" {
		server.simrobotNum = 100
	}

	server.robotId = 0
	server.simrobotId = 0

	server.maxId = 10000
	server.count = 50

	server.loginFinish = false

	//启动管理服务
	go server.run()

	//启动预登录
	go server.runPreLogin()

	if env == "dev" {
		//启动hua压力测试
		go server.runHua()

		//启动lhd压力测试
		go server.runLHD()

		//启动7up压力测试
		go server.runUP()

		//启动crash压力测试
		go server.runCRASH()
	}

	//启动测试
	//go server.runTest()
}

// Close 关闭连接
func (server *RobotServer) Close() {
	//关闭消息通道
	server.Send2rbs(closeFlag(1))
	server.remoteSend(closeFlag(1))

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}
	server.conns = nil
	server.mutexConns.Unlock()

	server.wg.Wait()
}

func (server *RobotServer) runRobot(id int, sim bool) {
	//host := getHost()
	//TODO test
	host := cfg.Section("gate.node1").Key("host").Value()

	var scheme string
	if env == "dev" {
		scheme = "ws"
	} else {
		scheme = "wss"
	}

	u := url.URL{Scheme: scheme, Host: host, Path: "/"}
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, _, err := dialer.Dial(u.String(), nil)

	if err != nil {
		glog.Errorf("robot run dial -> %v", err)
		return
	}

	server.wg.Add(1)
	defer server.wg.Done()

	server.mutexConns.Lock()
	if server.conns == nil {
		server.mutexConns.Unlock()
		conn.Close()
		return
	}
	server.conns[conn] = struct{}{}
	server.mutexConns.Unlock()

	//new robot
	robot := newRobot(conn, server.PendingWriteNum, server.MaxMsgLen)
	// robot.code = code //设置邀请码
	// robot.rtype = rtype
	// robot.ltype = ltype
	// robot.gtype = gtype
	// robot.gameid = gameid
	// robot.roomid = roomid
	// robot.envBet = envBet
	// robot.min = min
	// robot.max = max
	robot.sim = sim
	if sim {
		robot.data.Phone = fmt.Sprintf("simrobot%d", id)
	} else {
		robot.data.Phone = fmt.Sprintf("robot%d", id)
	}
	// robot.data.Phone = phone
	// robot.data.Nickname = phone
	// glog.Infof("run robot -> phone %s, roomid %s", phone, roomid)
	// glog.Infof("run robot -> code:%s, rtype:%d, regist:%v", code, rtype, regist)
	// regist = true //TODO test
	go robot.writePump()
	go robot.sendLogin()
	// if !regist {
	// 	go robot.sendRegist() //发起请求,注册-登录-进入房间
	// } else {
	// 	go robot.sendLogin() //登录
	// }
	go robot.ticker()
	go robot.pingPump()
	robot.readPump()

	// cleanup
	server.mutexConns.Lock()
	delete(server.conns, conn)
	server.mutexConns.Unlock()
}

func (r *RobotServer) getGtype() int32 {
	r.robotPoolMutex.Lock()
	defer r.robotPoolMutex.Unlock()
	for k, v := range r.robotNum {
		if v <= 0 {
			continue
		}
		if pool, ok := r.robotPool[k]; ok {
			if len(pool) >= v {
				continue
			}
		}
		return k
	}
	return 0
}

// RunRobot //' 启动一个机器人
// func (server *RobotServer) runRobot(gameid, roomid, phone, code string, rtype, ltype, gtype, envBet, min, max int32, sim, regist bool) {
// 	//host := getHost()
// 	//TODO test
// 	host := cfg.Section("gate.node1").Key("host").Value()
// 	u := url.URL{Scheme: "ws", Host: host, Path: "/"}
// 	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
// 	if err != nil {
// 		glog.Errorf("robot run dial -> %v", err)
// 		return
// 	}

// 	server.wg.Add(1)
// 	defer server.wg.Done()

// 	server.mutexConns.Lock()
// 	if server.conns == nil {
// 		server.mutexConns.Unlock()
// 		conn.Close()
// 		return
// 	}
// 	server.conns[conn] = struct{}{}
// 	server.mutexConns.Unlock()

// 	//new robot
// 	robot := newRobot(conn, server.PendingWriteNum, server.MaxMsgLen)
// 	robot.code = code //设置邀请码
// 	robot.rtype = rtype
// 	robot.ltype = ltype
// 	robot.gtype = gtype
// 	robot.gameid = gameid
// 	robot.roomid = roomid
// 	robot.envBet = envBet
// 	robot.min = min
// 	robot.max = max
// 	robot.sim = sim
// 	robot.data.Phone = phone
// 	robot.data.Nickname = phone
// 	glog.Infof("run robot -> phone %s, roomid %s", phone, roomid)
// 	glog.Infof("run robot -> code:%s, rtype:%d, regist:%v", code, rtype, regist)
// 	// regist = true //TODO test
// 	go robot.writePump()
// 	if !regist {
// 		go robot.sendRegist() //发起请求,注册-登录-进入房间
// 	} else {
// 		go robot.sendLogin() //登录
// 	}
// 	go robot.ticker()
// 	go robot.pingPump()
// 	robot.readPump()

// 	// cleanup
// 	server.mutexConns.Lock()
// 	delete(server.conns, conn)
// 	server.mutexConns.Unlock()
// }

// 获取网关
func getHost() (host string) {
	addr := cfg.Section("domain").Key("gate").Value()
	addr = "http://" + addr + "?version=" + *node
	b, err := doHttpPost(addr, []byte{})
	if err != nil {
		glog.Errorf("getHost err: %v", err)
		return
	}
	glog.Infof("getHost body: %s", string(b))
	host = string(b)
	if aesStatus {
		host = aesDe(b)
	}
	glog.Infof("getHost host: %s", host)
	return
}

//.

// doRequest get the order in json format with a sign
func doHttpPost(targetUrl string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("GET", targetUrl, bytes.NewBuffer([]byte(body)))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/x-www-form-urlencoded;charset=UTF-8")

	tr := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: false},
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return respData, nil
}

// vim: set foldmethod=marker foldmarker=//',//.:
