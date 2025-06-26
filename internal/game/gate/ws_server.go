package gate

import (
	"net"
	"net/http"
	"sync"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/glog"

	"github.com/gorilla/websocket"
)

type WSServer struct {
	Addr            string        //ws地址
	Addr2           string        //wss地址
	MaxConnNum      int           //最大连接数
	PendingWriteNum int           //等待写入消息长度
	MaxMsgLen       uint32        //最大消息长度
	HTTPTimeout     time.Duration //超时时间
	ln1             net.Listener  //监听
	ln2             net.Listener  //监听
	ln3             net.Listener  //监听
	ln4             net.Listener  //监听
	handler         *WSHandler    //处理
}

type WSHandler struct {
	maxConnNum      int                //最大连接数
	pendingWriteNum int                //等待写入消息长度
	maxMsgLen       uint32             //最大消息长
	upgrader        websocket.Upgrader //升级http连接
	conns           WebsocketConnSet   //连接集合
	mutexConns      sync.Mutex         //互斥锁
	wg              sync.WaitGroup     //同步机制
}

func (handler *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := handler.upgrader.Upgrade(w, r, nil)
	if err != nil {
		glog.Errorf("upgrade error: %v", err)
		return
	}
	conn.SetReadLimit(int64(handler.maxMsgLen))

	handler.wg.Add(1)
	defer handler.wg.Done()

	handler.mutexConns.Lock()
	if handler.conns == nil {
		handler.mutexConns.Unlock()
		conn.Close()
		return
	}
	if len(handler.conns) >= handler.maxConnNum {
		handler.mutexConns.Unlock()
		conn.Close()
		glog.Errorf("too many connections: %d", len(handler.conns))
		return
	}
	handler.conns[conn] = struct{}{}
	handler.mutexConns.Unlock()

	wsConn := newWSConn(conn, handler.pendingWriteNum, handler.maxMsgLen)
	wsConn.pid = wsConn.initWs()
	//start pid
	wsConn.pid.Tell(new(pb.ServeStart))
	go wsConn.writePump()
	go wsConn.pingPump()
	wsConn.readPump()

	// cleanup
	handler.mutexConns.Lock()
	delete(handler.conns, conn)
	handler.mutexConns.Unlock()
	// pid stop
	if wsConn.pid != nil {
		glog.Infof("wsConn.pid: %s", wsConn.pid.String())
		wsConn.pid.Tell(new(pb.ServeStop))
	}
}

func (server *WSServer) Start() {
	ln1, err := net.Listen("tcp4", "0.0.0.0"+server.Addr)
	if err != nil {
		glog.Fatal("%v", err)
	}

	ln2, err := net.Listen("tcp6", "[::]"+server.Addr)
	if err != nil {
		glog.Fatal("%v", err)
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 30000
		glog.Infof("invalid MaxConnNum, reset to %v", server.MaxConnNum)
	}
	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		glog.Infof("invalid PendingWriteNum, reset to %v", server.PendingWriteNum)
	}
	if server.MaxMsgLen <= 0 {
		server.MaxMsgLen = 8192 * 2
		glog.Infof("invalid MaxMsgLen, reset to %v", server.MaxMsgLen)
	}
	if server.HTTPTimeout <= 0 {
		server.HTTPTimeout = 10 * time.Second
		glog.Infof("invalid HTTPTimeout, reset to %v", server.HTTPTimeout)
	}

	server.ln1 = ln1
	server.ln2 = ln2
	server.handler = &WSHandler{
		maxConnNum:      server.MaxConnNum,
		pendingWriteNum: server.PendingWriteNum,
		maxMsgLen:       server.MaxMsgLen,
		conns:           make(WebsocketConnSet),
		upgrader: websocket.Upgrader{
			ReadBufferSize:   1024, //default 4096
			WriteBufferSize:  1024, //default 4096
			HandshakeTimeout: server.HTTPTimeout,
			CheckOrigin:      func(_ *http.Request) bool { return true },
		},
	}

	httpServer := &http.Server{
		// Addr:           server.Addr,
		Handler:        server.handler,
		ReadTimeout:    server.HTTPTimeout,
		WriteTimeout:   server.HTTPTimeout,
		MaxHeaderBytes: 1024,
	}

	go httpServer.Serve(ln1)
	go httpServer.Serve(ln2)

	if env == "pro" {
		certFile := "./myserver.pem"
		keyFile := "./myserver.key"

		ln3, err := net.Listen("tcp4", "0.0.0.0"+server.Addr2)
		if err != nil {
			glog.Fatal("%v", err)
		}

		ln4, err := net.Listen("tcp6", "[::]"+server.Addr2)
		if err != nil {
			glog.Fatal("%v", err)
		}

		httpServer2 := &http.Server{
			// Addr:           server.Addr,
			Handler:        server.handler,
			ReadTimeout:    server.HTTPTimeout,
			WriteTimeout:   server.HTTPTimeout,
			MaxHeaderBytes: 1024,
		}

		server.ln3 = ln3
		server.ln4 = ln4
		go httpServer2.ServeTLS(ln3, certFile, keyFile)
		go httpServer2.ServeTLS(ln4, certFile, keyFile)
		// go http.ListenAndServeTLS(server.Addr2, certFile, keyFile, server.handler)
	}

	// if env == "dev" {
	// 	go httpServer.Serve(ln)
	// } else {
	// 	certFile := "./myserver.pem"
	// 	keyFile := "./myserver.key"

	// 	go httpServer.ServeTLS(ln, certFile, keyFile)

	// 	go http.ListenAndServe("0.0.0.0:4108", server.handler)
	// }
}

func (server *WSServer) Close() {
	server.ln1.Close()
	server.ln2.Close()

	if env == "pro" {
		server.ln3.Close()
		server.ln4.Close()
	}

	server.handler.mutexConns.Lock()
	for conn := range server.handler.conns {
		conn.Close()
	}
	server.handler.conns = nil
	server.handler.mutexConns.Unlock()

	server.handler.wg.Wait()
}
