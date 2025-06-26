package login

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strconv"
	"strings"
)

// TouristLoginCheck 登录验证
func TouristLoginCheck(ctos *pb.TouristReq, key string) (stoc *pb.TouristRsp) {
	stoc = new(pb.TouristRsp)
	var account string = ctos.GetAccount()
	var passwd string = ctos.GetPassword()
	if len(passwd) != 32 {
		stoc.Error = pb.PwdFormatError
		return
	}
	//5秒失效
	n := utils.Timestamp() - (ctos.GetTime() / 1000)
	if n > 5 {
		glog.Debugf("tourist n %d, %d", n, ctos.GetTime())
		stoc.Error = pb.LoginError
		return
	}
	//解密验证
	account, err1 := TouristAccount(account, key)
	glog.Debugf("account %s, err1 %v", account, err1)
	if err1 != nil {
		stoc.Error = pb.LoginError
		return
	}
	if account == "" {
		stoc.Error = pb.LoginError
		return
	}
	//账号简单规则验证,TODO 优化
	str := utils.Split(account, "_")
	if len(str) != 2 {
		stoc.Error = pb.LoginError
		return
	}
	ran := utils.Split(str[1], ".")
	if len(ran) != 2 {
		stoc.Error = pb.LoginError
		return
	}
	//密码验证
	if passwd != utils.Md5(account) {
		stoc.Error = pb.LoginError
		return
	}
	ctos.Account = account
	return
}

// TouristLogin 游客登录
func TouristLogin(ctos *pb.TouristLogin, user *data.User) (stoc *pb.TouristLogined) {
	stoc = new(pb.TouristLogined)
	if user == nil {
		stoc.Error = pb.UsernameOrPwdError
		return
	}
	if user.Status == 2 || user.Status == 3 {
		stoc.Error = pb.AccountException
		return
	}
	// var passwd string = ctos.GetPassword()
	// if !user.VerifyPwd(passwd) {
	// 	glog.Errorf("Login error %s", user.GetUserid())
	// 	stoc.Error = pb.UsernameOrPwdError
	// }
	if user.Userid == "" {
		stoc.Error = pb.LoginError
	}
	if stoc.Error != pb.OK {
		return
	}
	stoc.Userid = user.Userid
	stoc.Appid = user.AD_AppId
	// setAfParam(user, ctos.Af)
	return
}

// TouristLoginRegist 游客注册
func TouristLoginRegist(arg *pb.TouristLogin, genid *data.IDGen, faceid *data.IDGen, count int) (stoc *pb.TouristLogined,
	user *data.User) {
	var account string = arg.GetAccount()
	var passwd string = arg.GetPassword()
	stoc = new(pb.TouristLogined)
	user = new(data.User)
	user.Tourist = account
	//if user.ExistsTourist() {
	//	stoc.Error = pb.PhoneRegisted
	//	user = nil
	//	return
	//}
	if count != 0 {
		stoc.Error = pb.RegistError
		glog.Errorf("device id is registed,%s", arg.Account)
		return
	}

	state := 1 // 默认新手状态
	stoc.Mode = 3
	// 检测设备码
	if arg.Ad != nil && arg.Ad.AdPlatform == "pc" {
		// pc端不能用游客
		count = 0
	}
	switch count {
	case 0:
		stoc.Mode = 2
	case 1:
		state = 3 // 同设备第二个号直接是平民
		stoc.Mode = 2
	}

	// photo := strconv.Itoa(utils.RandIntN(150) + 1)
	// photo := faceid.GenID()
	userid := genid.GenID()
	uid := data.GetUnusedUid()
	if uid != nil && uid.Uid != 0 {
		userid = fmt.Sprintf("%d", uid.Uid)
		uid.UpdateUid()
	}
	nickname := "Player" + userid
	glog.Debugf("TouristLogin userid %s", userid)
	auth := string(utils.GetAuth())

	token, _ := handler.Sign(userid)

	registArea := handler.UserABCType()
	// switch registArea {
	// case 1:
	// 	// B类直接是正常玩家
	// 	state = 2
	// 	stoc.Mode = 1
	// case 2:
	// 	// C类直接是正常玩家,什么也不送
	// 	state = 2
	// 	stoc.Mode = 3
	// }

	user = &data.User{
		Userid:   userid,
		Token:    token,
		Nickname: nickname,
		Auth:     auth,
		Password: utils.Md5(passwd + auth),
		Tourist:  account,
		Ctime:    utils.BsonNow(),
		// Photo:                         photo,
		Status:                        1,
		State:                         state,
		SignDay:                       1,
		WeekCardMap:                   make(map[string]*data.WeekCard),
		WithdrawLogMap:                make(map[string]*data.WithdrawLog),
		WithDrawCountMap:              make(map[int32]int32),
		FreeWinMap:                    make(map[int32][]data.FreeWin),
		Task:                          make(map[int32]*data.TaskInfo),
		ShareBelow:                    make(map[string]data.ShareData),
		TpUserFollowRateTrigger:       make(map[int32]int32),
		TpUserFollowRateSuccess:       make(map[int32]int32),
		TpUserStoryCD:                 make(map[int32]int32),
		TpUserTodayControlStrategyNum: make(map[int32]int32),
		RoundGames:                    make(map[int32]int32),
		GiftPopMap:                    make(map[string]data.GiftPop),
		RegistIP:                      arg.Registip,
		RegistMode:                    stoc.Mode,
		RegistArea:                    registArea,
		TpNewbieProbeId:               600, // tp新手试探局默认id
		// Diamond:          100000000,
		// Coin:             100000000,
		HistoryVersion: 1,
	}

	setAfParam(user, arg.Af)
	setADParam(user, arg.Ad)
	setFBParam(user, arg.Fb)

	//ad事件
	AdEvent(arg.Ad, user)

	// if arg.Af != nil {
	// 	user.DeviceId = arg.Af.Afid
	// 	user.OsVersion = arg.Af.Osversion
	// 	user.BundleId = arg.Af.Bundleid
	// 	user.AfId = arg.Af.Afid
	// 	user.MediaSource = arg.Af.Mediasource
	// 	user.AfKey = arg.Af.Afkey
	// }
	if stoc.Mode == 3 {
		user.State = 2
	}

	if !user.Save() {
		glog.Errorf("Regist save error %s", userid)
		stoc.Error = pb.RegistError
		return
	}
	if user.AD_RefGameId != "" {
		// 通过分享链接下载的
		shareStr := utils.Split(user.AD_RefGameId, "-")
		if len(shareStr) > 1 && shareStr[1] != "" {
			superId := shareStr[1]
			shareSource := 0
			if shares := strings.Split(superId, "_"); len(shares) > 1 {
				superId = shares[0]
				shareSource, _ = strconv.Atoi(shares[1])
			}
			user.ShareSuperior = superId
			user.ShareSource = int32(shareSource)
		}
	}
	stoc.Userid = user.Userid
	stoc.IsRegist = true
	stoc.Appid = user.AD_AppId
	return
}
