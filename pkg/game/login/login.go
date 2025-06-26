package login

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strings"
)

// LoginCheck 登录验证
func LoginCheck(ctos *pb.LoginReq) (stoc *pb.LoginRsp) {
	stoc = new(pb.LoginRsp)
	var phone string = ctos.GetPhone()
	var passwd string = ctos.GetPassword()
	if phone == "" {
		stoc.Error = pb.PhoneNumberEnpty
		return
	}
	if !utils.PhoneRegexp(phone) {
		glog.Errorf("PhoneRegexp error %s, %d", phone, len(phone))
		// stoc.Error = pb.PhoneNumberError
		//TODO 暂时不限制
		// return
	}
	if len(passwd) != 32 {
		stoc.Error = pb.PwdFormatError
	}
	return
}

// Login 登录
func Login(ctos *pb.RoleLogin, user *data.User) (stoc *pb.RoleLogined) {
	stoc = new(pb.RoleLogined)
	if user == nil {
		glog.Errorf("Login error %s no find user by phone %s", user.GetUserid(), ctos.Phone)
		stoc.Error = pb.UsernameOrPwdError
		return
	}
	if user.Status == 2 || user.Status == 3 {
		stoc.Error = pb.AccountException
		return
	}
	//var phone string = ctos.GetPhone()
	var passwd string = ctos.GetPassword()
	if !user.VerifyPwd(passwd) && user.Tourist == "" {
		glog.Errorf("Login error %s", user.GetUserid())
		stoc.Error = pb.UsernameOrPwdError
	}
	if user.Userid == "" {
		stoc.Error = pb.LoginError
	}
	if stoc.Error != pb.OK {
		return
	}
	stoc.Userid = user.Userid
	stoc.Status = int32(user.Status)
	stoc.Appid = user.AD_AppId
	// setAfParam(user, ctos.Af)
	return
}

// 设置归因参数
func setAfParam(user *data.User, param *pb.AfParam) {
	if param == nil {
		return
	}
	// user.DeviceId = param.GetDeviceid()
	user.AppId = param.GetAppId()
	user.AFId = param.GetAFId()
	user.OS = param.GetOS()
	user.BundleId = param.GetBundleId()
	user.RefGameId = param.GetRefGameId()
	user.RefPkgName = param.GetRefPkgName()
	user.MediaSource = param.GetMediaSource()
	user.AFKey = param.GetAFKey()
	user.DeviceId = param.GetDeviceId()
	user.Channel = param.GetChannel()
}

func setADParam(user *data.User, param *pb.ADParam) {
	if param == nil {
		return
	}
	user.AD_BundleId = param.BundleId
	user.AD_OsVersion = param.OsVersion
	user.AD_DeviceId = param.DeviceId
	user.AD_AppId = param.AppId
	user.AD_RefGameId = param.RefGameId
	user.AD_RefPkgName = param.RefPkgName
	user.AD_Channel = param.Channel
	user.AD_Tracker_Token = param.AdTrackerToken
	user.AD_Tracker_Name = param.AdTrackerName
	user.AD_Network = param.AdNetwork
	user.AD_Campaign = param.AdCampaign
	user.AD_ADGroup = param.AdAdgroup
	user.AD_Creative = param.AdCreative
	user.AD_Click_Lable = param.AdClickLabel
	user.AD_ADID = param.AdAdid
	user.AD_Cost_Type = param.AdCostType
	user.AD_Cost_Amount = param.AdCostAmount
	user.AD_Cost_Currency = param.AdCostCurrency
	user.AD_FB_Install_Referrer = param.AdFbInstallReferrer
	user.AD_Key = param.AdKey
	user.AD_S2S_Code = param.AdS2SCode
	user.AD_Event_Code1 = param.AdEventCode1
	user.AD_Event_Code2 = param.AdEventCode2
	user.AD_Event_Code3 = param.AdEventCode3
	user.AD_Event_Code4 = param.AdEventCode4
	user.AD_User_Agent = param.UserAgent
	user.AD_Tracker_Channel = param.AdTrackerChannel
	user.Platform = param.AdPlatform

	if user.AD_User_Agent == "" {
		user.AD_Fake_User_Agent = handler.GenUserAgent()
	}
}

func setFBParam(user *data.User, param *pb.FBParam) {
	if param == nil {
		return
	}
	if strings.HasPrefix("fb.", param.Fbc) {
		user.FB_Fbc = param.Fbc
	} else {
		user.FB_Fbc = fmt.Sprintf("fb.1.%d.%s", utils.BsonNow().UnixMilli(), param.Fbc)
	}
	user.FB_Fbp = param.Fbp
}
