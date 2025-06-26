package robot

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"runtime/debug"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/gogo/protobuf/proto"
)

// RoleActor 玩家角色进程
type RoleActor struct {
	stopCh chan struct{} // 关闭通道

	pid     *actor.PID // rs进程ID
	dbmsPid *actor.PID // 数据中心
	roomPid *actor.PID // 房间节点
	gamePid *actor.PID // 游戏逻辑
	gameId  string     //游戏id
	roomId  string     //房间id
	gtype   int32      //游戏类型

	*data.User      //玩家在线数据
	*DeskData       //桌子数据
	isReal     bool //是否使用真实头像

	online bool //在线状态
	status bool //更新状态
	timer  int  //计时
}

// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
func (a *RoleActor) Receive(ctx actor.Context) {
	defer func() {
		if r := recover(); r != nil {
			glog.Error("Receive handler recover error:", r)
			debug.PrintStack()
		}
	}()
	switch msg := ctx.Message().(type) {
	case *pb.Request:
		ctx.Respond(&pb.Response{})
	case *actor.Started:
		glog.Notice("Starting, initialize actor here")
	case *actor.Stopping:
		glog.Notice("Stopping, actor is about to shut down")
	case *actor.Stopped:
		glog.Notice("Stopped, actor and its children are stopped")
	case *actor.Restarting:
		glog.Notice("Restarting, actor is about to restart")
	case *actor.ReceiveTimeout:
		glog.Infof("ReceiveTimeout: %v", ctx.Self().String())
	case proto.Message:
		a.Handler(msg, ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// NewRole 启动一个新玩家角色
func NewRole() *RoleActor {
	return &RoleActor{
		stopCh: make(chan struct{}),
	}
}

// 初始化
func (rs *RoleActor) initRs() *actor.PID {
	props := actor.FromProducer(func() actor.Actor { return rs }) //实例
	return actor.Spawn(props)                                     //启动一个进程
}
