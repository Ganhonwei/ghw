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

	"github.com/globalsign/mgo/bson"
)

// RestPwdCheck 重置密码验证
func RestPwdCheck(ctos *pb.ResetPwdReq) (stoc *pb.ResetPwdRsp) {
	stoc = new(pb.ResetPwdRsp)
	var phone string = ctos.GetPhone()
	var passwd string = ctos.GetPassword()
	var smscode string = ctos.GetSmscode()
	if len(smscode) != 6 {
		stoc.Error = pb.SmsCodeEmpty
		return
	}
	if !utils.PhoneValidate(phone) {
		glog.Errorf("PhoneValidate error %s", phone)
		stoc.Error = pb.PhoneNumberError
		return
	}
	if len(passwd) != 32 {
		stoc.Error = pb.PwdFormatError
	}
	return
}

// RegistCheck 注册验证
func RegistCheck(ctos *pb.RegistReq) (stoc *pb.RegistRsp) {
	stoc = new(pb.RegistRsp)
	var nickname string = ctos.GetNickname()
	var phone string = ctos.GetPhone()
	var passwd string = ctos.GetPassword()
	var smscode string = ctos.GetSmscode()
	var safetycode string = ctos.GetSafetycode()
	if phone == "" {
		stoc.Error = pb.PhoneNumberEnpty
		return
	}
	if nickname == "" {
		stoc.Error = pb.UsernameEmpty
		return
	}
	if len(safetycode) == 0 {
		// stoc.Error = pb.SafetycodeEmpty
		// return
	}
	if len(smscode) != 6 {
		//stoc.Error = pb.SmsCodeEmpty
		//TODO 暂时不限制
		//return
	}
	if !utils.LegalName(nickname, 7) {
		glog.Errorf("LegalName error %s", nickname)
		stoc.Error = pb.NameTooLong
		return
	}
	if !utils.PhoneValidate(phone) {
		glog.Errorf("PhoneValidate error %s", phone)
		//stoc.Error = pb.PhoneNumberError
		//TODO 暂时不限制
		//return
	}
	if len(passwd) != 32 {
		stoc.Error = pb.PwdFormatError
	}
	return
}

// Regist 注册处理
func Regist(arg *pb.RoleRegist, genid *data.IDGen, faceid *data.IDGen) (stoc *pb.RoleRegisted,
	user *data.User) {
	var nickname string = arg.GetNickname()
	var phone string = arg.GetPhone()
	var passwd string = arg.GetPassword()
	// var safetycode string = arg.GetSafetycode()
	stoc = new(pb.RoleRegisted)
	stoc.Mode = 3
	state := 1
	// 检测设备码
	if !strings.Contains(phone, "robot") && !strings.Contains(phone, "simrobot") {
		// 非人机才检测
		if arg.Ad == nil {
			stoc.Error = pb.RegistError
			return
		} else {
			count := data.Count(data.PlayerUsers, bson.M{"ad__device_id": arg.Ad.DeviceId})
			if arg.Ad.AdPlatform == "pc" {
				// pc端没有限制,直接是新手
				count = 0
			}
			switch count {
			case 0:
				stoc.Mode = 2
			// 一个设备注册一个账号
			// case 1:
			// 	state = 3 // 同设备第二个号直接是平民
			// 	stoc.Mode = 2
			default:
				stoc.Error = pb.DeviceUnableRegist
				return
			}
		}
	}

	userid := genid.GenID()
	uid := data.GetUnusedUid()
	if uid != nil && uid.Uid != 0 {
		userid = fmt.Sprintf("%d", uid.Uid)
		uid.UpdateUid()
	}
	// photo := strconv.Itoa(utils.RandIntN(150) + 1)
	// photo := faceid.GenID()
	glog.Debugf("RoleRegist userid %s", userid)
	auth := string(utils.GetAuth())
	nickname = "Player" + userid

	var robot bool = false
	var simrobot bool = false

	if strings.HasPrefix(phone, "robot") {
		robot = true
		nickname = RandName()
	} else if strings.HasPrefix(phone, "simrobot") {
		simrobot = true
		nickname = "Sim" + userid
	}

	token, _ := handler.Sign(userid)

	registArea := handler.UserABCType()
	// switch registArea {
	// case 1:
	// 	// B类直接是正常玩家
	// 	state = 2
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
		// Agent:       safetycode,
		// Atime:       utils.BsonNow(),
		Password: utils.Md5(passwd + auth),
		Phone:    phone,
		// Photo:    photo,
		Ctime: utils.BsonNow(),
		// DeviceId:         arg.Af.GetAfid(),
		// OsVersion:        arg.Af.GetOsversion(),
		// BundleId:         arg.Af.GetBundleid(),
		// AfId:             arg.Af.GetAfid(),
		// MediaSource:      arg.Af.GetMediasource(),
		// AfKey:            arg.Af.GetAfkey(),
		Status:           1,
		State:            state,
		SignDay:          1,
		WeekCardMap:      make(map[string]*data.WeekCard),
		WeeklyCardMap:    make(map[int32]*data.WeekCard),
		WithdrawLogMap:   make(map[string]*data.WithdrawLog),
		WithDrawCountMap: make(map[int32]int32),
		Task:             make(map[int32]*data.TaskInfo),
		FreeWinMap:       make(map[int32][]data.FreeWin),
		ShareBelow:       make(map[string]data.ShareData),
		RoundGames:       make(map[int32]int32),
		GiftPopMap:       make(map[string]data.GiftPop),
		Robot:            robot,
		SimRobot:         simrobot,
		RegistIP:         arg.Ip,
		RegistMode:       stoc.Mode,
		RegistArea:       registArea,
		TpNewbieProbeId:  600, // tp新手试探局默认id
		HistoryVersion:   1,
	}
	setAfParam(user, arg.Af)
	setADParam(user, arg.Ad)
	setFBParam(user, arg.Fb)

	//ad事件
	AdEvent(arg.Ad, user)

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
	return
}

// RobotRegist Robot注册处理
func RobotRegist(arg *pb.RobotRegist, genid *data.IDGen) {
	user := new(data.User)
	user.Phone = arg.GetPhone()
	user.GetByPhone() //数据库中取
	if user.Userid != "" {
		glog.Debugf("account %s exist", arg.GetPhone())
		return
	}
	userid := genid.GenID()
	glog.Debugf("RobotRegist userid %s", userid)
	user.Userid = userid
	user.Nickname = arg.GetNickname()
	user.Phone = arg.GetPhone()
	user.Photo = arg.GetPhoto()
	user.Password = utils.Md5(arg.GetPassword() + arg.GetAuth())
	user.Auth = arg.GetAuth()
	user.Sex = arg.GetSex()
	user.Coin = arg.GetCoin()
	user.Diamond = arg.GetDiamond()
	user.Ctime = utils.BsonNow()
	user.Robot = true
	if !user.Save() {
		glog.Errorf("RobotRegist save failed %#v, userid %s", arg, userid)
		return
	}
	glog.Debugf("RobotRegist successfully userid %s, phone %s", userid, arg.GetPhone())
}

// ad事件
func AdEvent(ad *pb.ADParam, user *data.User) {
	if ad == nil {
		glog.Errorf("user:%s adparam is nil", user.Userid)
		return
	}
	e := new(data.AdjustEvent)
	data.Get(data.ADJusts, ad.AppId, &e)
	if e.AppId == "" {
		return
	}
	user.AD_AppId = ad.AppId
	user.AD_Key = e.AdKey
	user.AD_S2S_Code = e.AdS2sCode
	user.AD_Event_Code1 = e.AdEventCode1
	user.AD_Event_Code2 = e.AdEventCode2
	user.AD_Event_Code3 = e.AdEventCode3
	user.AD_Event_Code4 = e.AdEventCode4
	user.AD_Event_Code5 = e.AdEventCode5
}
