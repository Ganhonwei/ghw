package pay

import (
	"goserver/gen/pb"
	"goserver/pkg/glog"
	"net/url"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (a *PayActor) Handler(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *pb.ServeStart:
		a.ServeStart(ctx)
	case *pb.ServeStop:
		a.ServeStop(ctx)
	case *pb.Tick:
		a.ding(ctx)
	default:
		glog.Errorf("unknown message %v", msg)
	}
}

// 启动服务
func (a *PayActor) ServeStart(ctx actor.Context) {
	glog.Infof("pay start: %v", ctx.Self().String())
	//启动
	go a.ticker(ctx)
}

// 停止服务
func (a *PayActor) ServeStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//TODO 关闭消息
	//响应
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

// 关闭时钟
func (a *PayActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *PayActor) ticker(ctx actor.Context) {
	tick := time.Tick(2 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("desk ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 滴答
func (a *PayActor) ding(ctx actor.Context) {
	switch a.timer {
	case 300:
		// 查询10-20分钟前的订单.
		go InspectOrder()
		a.timer = 0
	default:
		a.timer++
	}
}

func callNode(msg interface{}) (interface{}, error) {
	timeout := 5 * time.Second
	res, err := nodePid.RequestFuture(msg, timeout).Result()
	return res, err
}

// URLENCODE编码转换
func ToUrlEncode(str string) string {
	encodedURL := url.QueryEscape(str)
	return encodedURL
}
