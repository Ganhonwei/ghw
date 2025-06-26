package dbms

import (
	"context"
	"fmt"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/login"
	"goserver/pkg/glog"
	"goserver/pkg/myactor"
	"goserver/pkg/utils"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/globalsign/mgo/bson"
	"github.com/redis/go-redis/v9"
)

func (a *RoleActor) Connected(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Connected)
	glog.Infof("Connected %s", arg.Name)
}

func (a *RoleActor) Disconnected(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.Disconnected)
	glog.Infof("Disconnected %s", arg.Name)
}

func (a *RoleActor) ServeStop(ctx actor.Context) {
	glog.Debugf("handlerStop: %s", a.Name)
	//关闭
	a.closeTick()
	//回存数据
	if a.uniqueid != nil {
		a.uniqueid.Save()
	}
	for k, v := range a.offline {
		glog.Debugf("Stop offline: %s", k)
		v.Save()
	}
	for k, v := range a.online {
		glog.Debugf("Stop online: %s", k)
		v.Save()
	}
	//响应登录
	rsp := new(pb.ServeStoped)
	ctx.Respond(rsp)
}

func (a *RoleActor) InitRobot() {
	count := 10000

	passwd := cfg.Section("robot").Key("passwd").Value()
	passwd = utils.Md5(passwd)

	robot_count := data.Count(data.PlayerUsers, bson.M{"phone": bson.M{"$regex": "^robot"}})
	robot_lack := count - robot_count

	if robot_lack > 0 {
		robot_id := robot_count + 1
		for i := 0; i < robot_lack; i++ {
			login.Regist(&pb.RoleRegist{Phone: fmt.Sprintf("robot%d", robot_id), Password: passwd}, a.uniqueid, a.faceid)
			robot_id += 1
		}
	}
}

// 初始化人机头像
func (a *RoleActor) InitRobotHead() {
	on := cfg.Section("env").Key("initRobotHead").MustBool(false)
	if !on {
		return
	}
	//建立redis连接
	redisAddr := cfg.Section("redis").Key("addr").Value()
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       3,
	})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}

	var manEnNames = make([]string, 0)
	var manInNames = make([]string, 0)
	var womanEnNames = make([]string, 0)
	var womanInNames = make([]string, 0)
	var manPhotos = make([]string, 0)
	var womanPhotos = make([]string, 0)
	_ = manPhotos
	_ = womanPhotos

	initRobotNames := func() {
		file, err := excelize.OpenFile("config/name.xlsx")
		if err != nil {
			panic(err)
		}
		var err1, err2, err3 error
		manEnNames, manInNames, err1 = config.GetRobotNameByExcel("man", file)
		if err1 != nil {
			panic(err1)
		}
		womanEnNames, womanInNames, err2 = config.GetRobotNameByExcel("woman", file)
		if err2 != nil {
			panic(err2)
		}
		manPhotos, womanPhotos, err3 = config.GetRobotNameByExcel("photo", file)
		if err3 != nil {
			panic(err3)
		}
	}

	initRobotNames()

	getNameAndSex := func() (name string, sex uint32) {
		// 50%概率使用配置的名字
		india := false
		if utils.RandWan(2000) {
			// 剩下的人中20%使用印度名字
			india = true
		}
		if utils.RandWan(6000) {
			// 剩下的人中60%中是男性
			sex = 1
			// i := utils.RandInt32N(int32(len(manPhotos)))
			// photo = manPhotos[i]
			if india && len(manInNames) > 0 {
				index := utils.RandInt32N(int32(len(manInNames)))
				name = manInNames[index]
				manInNames = append(manInNames[:index], manInNames[index+1:]...)
			} else {
				index := utils.RandInt32N(int32(len(manEnNames)))
				name = manEnNames[index]
				manEnNames = append(manEnNames[:index], manEnNames[index+1:]...)
			}
		} else {
			// 女性
			sex = 2
			// i := utils.RandInt32N(int32(len(womanPhotos)))
			// photo = womanPhotos[i]
			if india {
				index := utils.RandInt32N(int32(len(womanInNames)))
				name = womanInNames[index]
				womanInNames = append(womanInNames[:index], womanInNames[index+1:]...)
			} else {
				index := utils.RandInt32N(int32(len(womanEnNames)))
				name = womanEnNames[index]
				womanEnNames = append(womanEnNames[:index], womanEnNames[index+1:]...)
			}
		}
		return
	}

	//初始化对象池
	man := 7211
	woman := 809

	type HeadInfo struct {
		Id   uint32
		Name string
		Sex  uint32
	}

	initPool := func(count int, poolName string) error {
		for i := 0; i < count; i++ {
			name, sex := getNameAndSex()
			info := HeadInfo{
				Id:   uint32(i),
				Name: name,
				Sex:  sex,
			}
			b, err := json.Marshal(info)
			if err != nil {
				return err
			}

			err = redisClient.HSet(context.Background(), fmt.Sprintf("%s:hash", poolName), fmt.Sprintf("%d", i), string(b)).Err()
			if err != nil {
				return err
			}

			err = redisClient.SAdd(context.Background(), fmt.Sprintf("%s:set", poolName), fmt.Sprintf("%d", i)).Err()
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = initPool(man, "man")
	if err != nil {
		panic(err)
	}
	err = initPool(woman, "woman")
	if err != nil {
		panic(err)
	}
}

// 初始化人机信息
func (a *RoleActor) InitRobotInfo() {
	on := cfg.Section("env").Key("initRobotInfo").MustBool(false)
	if !on {
		return
	}
	// file, err := excelize.OpenFile("name.xlsx")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer file.Close()
	// // 男性名字
	// manEnNames, manInNames, err1 := a.GetNameByExcel("man", file)
	// if err1 != nil {
	// 	fmt.Println(err1)
	// 	return
	// }
	// // 女性名字
	// womanEnNames, womanInNames, err2 := a.GetNameByExcel("woman", file)
	// if err2 != nil {
	// 	fmt.Println(err2)
	// 	return
	// }
	// // 头像
	// manPhotos, womanPhotos, err3 := a.GetNameByExcel("photo", file)
	// if err3 != nil {
	// 	fmt.Println(err3)
	// 	return
	// }
	// userids := make([]data.RobotUserId, 0)
	// data.ListByQWithFields(data.PlayerUsers, bson.M{"phone": bson.M{"$regex": "^robot"}}, []string{"_id"}, &userids)
	// for _, v := range userids {
	// 	uid := data.GetUnusedUid()
	// 	name := fmt.Sprintf("Player%s", v.ID)
	// 	photo := fmt.Sprintf("%d", utils.RandInt32N(150)+1)
	// 	var sex uint32 = uint32(utils.RandInt32N(2) + 1)
	// 	if uid != nil {
	// 		name = fmt.Sprintf("Player%d", uid.Uid)
	// 		uid.UpdateUid()
	// 	}
	// 	if utils.RandWan(5000) {
	// 		// 50%概率使用配置的名字
	// 		india := false
	// 		if utils.RandWan(2000) {
	// 			// 剩下的人中20%使用印度名字
	// 			india = true
	// 		}
	// 		if utils.RandWan(6000) {
	// 			// 剩下的人中60%中是男性
	// 			sex = 1
	// 			i := utils.RandInt32N(int32(len(manPhotos)))
	// 			photo = manPhotos[i]
	// 			if india && len(manInNames) > 0 {
	// 				index := utils.RandInt32N(int32(len(manInNames)))
	// 				name = manInNames[index]
	// 				manInNames = append(manInNames[:index], manInNames[index+1:]...)
	// 			} else {
	// 				index := utils.RandInt32N(int32(len(manEnNames)))
	// 				name = manEnNames[index]
	// 				manEnNames = append(manEnNames[:index], manEnNames[index+1:]...)
	// 			}
	// 		} else {
	// 			// 女性
	// 			sex = 2
	// 			i := utils.RandInt32N(int32(len(womanPhotos)))
	// 			photo = womanPhotos[i]
	// 			if india {
	// 				index := utils.RandInt32N(int32(len(womanInNames)))
	// 				name = womanInNames[index]
	// 				womanInNames = append(womanInNames[:index], womanInNames[index+1:]...)
	// 			} else {
	// 				index := utils.RandInt32N(int32(len(womanEnNames)))
	// 				name = womanEnNames[index]
	// 				womanEnNames = append(womanEnNames[:index], womanEnNames[index+1:]...)
	// 			}
	// 		}
	// 	}
	// 	data.Update(data.PlayerUsers, bson.M{"_id": v.ID}, bson.M{"$set": bson.M{"nickname": name, "photo": photo, "sex": sex}})
	// }
}

func (a *RoleActor) GetNameByExcel(sheet string, f *excelize.File) ([]string, []string, error) {
	manRows, err1 := f.GetRows(sheet)
	if err1 != nil {
		fmt.Println(err1)
		return nil, nil, err1
	}
	manEnNames := make([]string, 0) // 英文名
	manInNames := make([]string, 0) // 印度名
	for _, v := range manRows {
		if v[0] != "" {
			manEnNames = append(manEnNames, v[0])
		}
		if len(v) > 1 && v[1] != "" {
			manInNames = append(manInNames, v[1])
		}
	}
	return manEnNames, manInNames, nil
}

func (a *RoleActor) InitSimRobot() {
	count := 10000

	passwd := cfg.Section("robot").Key("passwd").Value()
	passwd = utils.Md5(passwd)

	robot_count := data.Count(data.PlayerUsers, bson.M{"phone": bson.M{"$regex": "^simrobot"}})
	robot_lack := count - robot_count

	if robot_lack > 0 {
		robot_id := robot_count + 1
		for i := 0; i < robot_lack; i++ {
			login.Regist(&pb.RoleRegist{Phone: fmt.Sprintf("simrobot%d", robot_id), Password: passwd}, a.uniqueid, a.faceid)
			robot_id += 1
		}
	}
}

// 启动服务
func (a *RoleActor) ServeStart(ctx actor.Context) {
	glog.Infof("role start: %v", ctx.Self().String())
	a.InitRobotHead()
	// a.InitRobot()
	// a.InitRobotInfo()
	// a.InitSimRobot()
	////注册更新机器人
	//phone := cfg.Section("robot").Key("phone").Value()
	//passwd := cfg.Section("robot").Key("passwd").Value()
	////data.RegistRobots(phone, passwd, a.uniqueid)
	////新机器人
	//head := cfg.Section("domain").Key("headimag").Value()
	////data.RegistRobots2(head, phone, passwd, a.uniqueid)
	//data.RegistRobots3(head, phone, passwd, a.uniqueid)
	//启动
	go a.ticker(ctx)
	go a.ticker2(ctx)
	go a.ticker3(ctx)
	//响应
	//rsp := new(pb.ServeStarted)
	//ctx.Respond(rsp)

	// 初始化小米抽奖活动
	go a.initLuckyDraw()

	// 初始化转盘活动
	a.ActivityTurnInit()

	consumerName := "dbms-role"
	// 新用户邀请订阅
	instShareSuperior := func() *pb.PublishShareRegist { return new(pb.PublishShareRegist) }
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicShareSuperior, instShareSuperior, func(msg *pb.PublishShareRegist) (err error) {
		// 查询设备码是否存在
		existsAdid := data.ExistsAdid(msg.Adid, msg.Userid)
		// 转盘活动加次数
		rolePid.Tell(&pb.ActivityTurnShareRegist{
			Userid:        msg.Userid,
			ShareSuperior: msg.ShareSuperior,
			NewAdid:       !existsAdid,
		})
		// 代理注册消息
		rolePid.Tell(&pb.ShareAgentRegist{Userid: msg.Userid, SuperId: msg.ShareSuperior})
		return
	})

	// 打码上报订阅
	instGameBets := func() *pb.PublishGameBets { return new(pb.PublishGameBets) }
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameBets, instGameBets, func(msg *pb.PublishGameBets) (err error) {
		// 代理打码
		if !msg.Robot && msg.SuperId != "" {
			rolePid.Tell(&pb.ShareAgentBets{
				Userid:  msg.Userid,
				SuperId: msg.SuperId,
				Bets:    msg.Bets,
				Gtype:   msg.Gtype,
				WaterId: msg.WaterId,
			})
		}

		// 波动返水打码
		if !msg.Robot {
			volatilitySubsidyBet(msg.Userid, msg.RegistArea, msg.Bets)
			// 波动返水返奖
			if msg.Score > 0 {
				volatilitySubsidyReward(msg.Userid, msg.RegistArea, msg.Score)
			}
		}
		return nil
	}, mq.WithConsumerNoneAck(), mq.WithConsumerMaxDeliver(1))

	// 返回大厅订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameToLobby,
		func() *pb.PublishGameToLobby { return new(pb.PublishGameToLobby) },
		func(msg *pb.PublishGameToLobby) (err error) {
			if !msg.Login {
				// 波动返水判断弹窗
				rolePid.Tell(&pb.VolatilitySubsidyToLobby{Userid: msg.Userid})
				// 活动弹窗
				a.ActivityPop(msg.Userid, 2)
			}
			return
		})

	// 外接订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicExternalBet,
		func() *pb.PublishExternalBet { return new(pb.PublishExternalBet) }, handleExternalBet)
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicExternalReword,
		func() *pb.PublishExternalReward { return new(pb.PublishExternalReward) }, handleExternalReward)
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicExternalCancel,
		func() *pb.PublishExternalCancel { return new(pb.PublishExternalCancel) }, handleExternalCancel)

	// 排行榜弹窗订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamActivity, mq.TopicActivityBetRankWindow, func() *pb.ActivityBetRankWindowNtfs { return new(pb.ActivityBetRankWindowNtfs) },
		func(msg *pb.ActivityBetRankWindowNtfs) (err error) {
			rolePid.Tell(msg)
			return
		})

	// UTR结果订阅
	mq.NatsCreateConsumer(consumerName, mq.StreamUtr, mq.TopicUtrUploadResult, func() *pb.UploadUtrRsp { return new(pb.UploadUtrRsp) },
		func(msg *pb.UploadUtrRsp) (err error) {
			rolePid.Tell(msg)
			return
		})

	// 优惠券配置
	mq.NatsCreateConsumer(consumerName, mq.StreamActivity, mq.TopicActivityCouponUpdate, func() *pb.ManualCouponUpdate { return new(pb.ManualCouponUpdate) }, func(msg *pb.ManualCouponUpdate) (err error) {
		coupon := data.GetManualCouponById(msg.CouponId)
		if coupon == nil {
			glog.Errorf("coupon not found, couponId: %s", msg.CouponId)
			return
		}
		config.SetManualCoupon(coupon)
		return
	})
	// 用户登录
	mq.NatsCreateConsumer(consumerName, mq.StreamGame, mq.TopicGameLogin, func() *pb.PublishGameLogin { return new(pb.PublishGameLogin) }, func(msg *pb.PublishGameLogin) (err error) {
		// 弹窗判断
		a.ActivityPop(msg.Userid, 1)
		return
	})
	// 提现转单
	mq.NatsCreateConsumer(consumerName, mq.StreamWithdraw, mq.TopicWithdrawTransfer, func() *pb.WithdrawTransfer { return new(pb.WithdrawTransfer) }, func(msg *pb.WithdrawTransfer) (err error) {
		rolePid.Tell(msg)
		return
	})
}

// 时钟
func (a *RoleActor) ticker(ctx actor.Context) {
	tick := time.Tick(30 * time.Second)
	msg := new(pb.Tick)
	for {
		select {
		case <-a.stopCh:
			glog.Info("role ticker closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("role ticker closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 时钟
func (a *RoleActor) ticker2(ctx actor.Context) {
	tick := time.Tick(2 * time.Second)
	msg := new(pb.Tick2)
	for {
		select {
		case <-a.stopCh:
			glog.Info("role ticker2 closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("role ticker2 closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 时钟
func (a *RoleActor) ticker3(ctx actor.Context) {
	tick := time.Tick(1 * time.Minute)
	msg := new(pb.Tick3)
	for {
		select {
		case <-a.stopCh:
			glog.Info("role ticker3 closed")
			return
		default: //防止阻塞
		}
		select {
		case <-a.stopCh:
			glog.Info("role ticker3 closed")
			return
		case <-tick:
			ctx.Self().Tell(msg)
		}
	}
}

// 钟声
func (a *RoleActor) Tick(ctx actor.Context) {
	//glog.Debugf("ding: %v", ctx.Self().String())
	//glog.Debugf("timer: %d", a.timer)
	switch a.timer {
	case 4: //2分钟
		a.saveUser()
		a.smsExpire()
		a.touristIP()
	case 2: //1分钟
		a.smsExpire()
		a.timer++
	default:
		a.timer++
	}
	switch a.shareTimer {
	case 120:
		if env == "dev" {
			go a.shareSettlement()
		} else {
			// 分享活动结算早上9点
			hour := utils.LocalTime().Hour()
			if hour == 9 {
				go a.shareSettlement()
			}
		}
		a.shareTimer = 0
	default:
		a.shareTimer++
	}
	// 跑马灯
	a.marquee()
	// 充值检测
	a.checkCharge()

	// 记录每分钟在线人数
	now := time.Now().In(location)
	if curMinute := now.Minute(); a.lastLogOnlineUsersMinute != curMinute {
		a.OnlineUsersLog(now)
		a.lastLogOnlineUsersMinute = curMinute
		now2 := time.Now().In(location)
		glog.Infof("tick online users log: use %dms", now2.UnixMilli()-now.UnixMilli())
	}
}

// 2s tick
func (a *RoleActor) Tick2(ctx actor.Context) {
	// 小米手机活动奖券放号
	// a.luckyDrawTime()
	// 新跑马灯(关闭)
	// a.marqueeWithdrawTime()
	// 转盘活动时钟
	a.activityTurnTick()
	// 波动返水活动
	a.volatilitySubsidyTick()
}

func (a *RoleActor) Tick3(ctx actor.Context) {

}

// 关闭时钟
func (a *RoleActor) closeTick() {
	select {
	case <-a.stopCh:
		return
	default:
		//停止发送消息
		close(a.stopCh)
	}
}

func (a *RoleActor) CloseServer(ctx actor.Context) {
	a.serverClose = true
}

func (a *RoleActor) checkCharge() {
	now := utils.BsonNow().Unix()
	if a.lastChargeTime+3600 > now {
		return
	}
	glog.Infof("check for an hour without charging")

	a.lastChargeTime = now
	// 一个小时没有充值
	{
		event := &pb.AlertorMail{Subject: "System Msg", Message: "No Charge For An Hour"}
		// body, err := event.Marshal()
		// if err != nil {
		// 	glog.Errorf("marshal err: %v", err)
		// 	return
		// }

		// err = producer.Publish(data.TopicAlertor, body)
		err = mq.NatsPublish(mq.TopicAlertEmail, event)
		if err != nil {
			glog.Errorf("publish alertor err: %v", err)
			return
		}
	}
}

// 记录每分钟在线人数
func (a *RoleActor) OnlineUsersLog(now time.Time) {
	log := &pb.OnlineUsersLog{}
	rsp := &pb.OnlineUserMinuteRsp{}
	err := mq.NatsRequest(mq.RequestOnlineUserCount, rsp, &mq.RequestEmptyArgs{})
	if err != nil {
		glog.Errorf("online users log err: %v", err)
		return
	}

	// log.Users = uint32(len(a.offline) + len(a.online))
	log.Users = uint32(len(rsp.UserIds))
	if log.Users == 0 {
		return
	}

	log.Ctime = now.Unix()
	log.DayTime = uint32(now.Year()*10000 + int(now.Month())*100 + now.Day())
	log.MinuteTime = uint32(now.Hour()*100 + now.Minute())
	log.Id = int64(log.DayTime)*10000 + int64(log.MinuteTime)
	for _, userid := range rsp.UserIds {
		user := a.getUserById(userid)
		if user == nil {
			continue
		}
		log.OnlineUsers++
		if user.Money > 0 {
			// log.OnlinePayUsers++
			log.PayUsers++
		}
	}

	// for _, user := range a.online {
	// 	req := new(pb.OnlineUser)
	// 	online := false
	// 	// 获取玩家当前位置判断是否在游戏中
	// 	if role, ok := a.roles[user.Userid]; ok {
	// 		body, err := role.Pid.RequestFuture(req, 50*time.Millisecond).Result()
	// 		if err != nil {
	// 			glog.Errorf("[onlineuser] user: %s, err: %v", err, user.Userid)
	// 		} else {
	// 			if r, ok := body.(*pb.OnlinedUser); ok {
	// 				online = r.Online
	// 			}
	// 		}
	// 	}

	// 	if online {
	// 		log.OnlineUsers++
	// 		if user.Money > 0 {
	// 			log.OnlinePayUsers++
	// 			log.PayUsers++
	// 		}
	// 	} else {
	// 		log.OfflineUsers++
	// 		if user.Money > 0 {
	// 			log.OfflinePayUsers++
	// 			log.PayUsers++
	// 		}
	// 	}
	// }

	// for _, user := range a.offline {
	// 	log.OfflineUsers++
	// 	if user.Money > 0 {
	// 		log.OfflinePayUsers++
	// 		log.PayUsers++
	// 	}
	// }
	myactor.Logger().Tell(log)
}
