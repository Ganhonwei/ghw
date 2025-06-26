package login

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
)

func ReportRegisterMsg(user *data.User) *pb.ReportRegister {
	msg := &pb.ReportRegister{}

	// if user.AppId != "" && user.AFId != "" && user.OS != "" && user.BundleId != "" && user.MediaSource != "" && user.AFKey != "" {
	// 	msg.AfInfo = &pb.AFInfo{
	// 		APPID:       user.AppId,
	// 		AFID:        user.AFId,
	// 		OS:          user.OS,
	// 		BundleID:    user.BundleId,
	// 		RefGameID:   user.RefGameId,
	// 		RefPkgName:  user.RefPkgName,
	// 		MediaSource: user.MediaSource,
	// 		AFKey:       user.AFKey,
	// 	}
	// }

	// appid中不包含.的都要上报到fb
	// if !strings.Contains(user.AD_AppId, ".") {
	// 	// 上报fb
	// 	msg.FbInfo = &pb.FBInfo{
	// 		FBId: ,
	// 	}
	// }

	if user.AD_Key != "" && user.AD_S2S_Code != "" && user.AD_Event_Code1 != "" && user.AD_Event_Code2 != "" && user.AD_Event_Code3 != "" && user.AD_Event_Code4 != "" {
		msg.AdInfo = &pb.ADInfo{
			ADID:         user.AD_ADID,
			ADKey:        user.AD_Key,
			ADS2SCode:    user.AD_S2S_Code,
			ADEventCode1: user.AD_Event_Code1,
			ADEventCode2: user.AD_Event_Code2,
			ADEventCode3: user.AD_Event_Code3,
			ADEventCode4: user.AD_Event_Code4,
			UserAgent:    user.AD_User_Agent,
			Ip:           user.RegistIP,
			BundleID:     user.AD_BundleId,
			Userid:       user.Userid,
		}
	}
	if user.FB_Fbc != "" {
		msg.FbInfo = new(pb.FBInfo)
	}

	return msg
}

// func ReportRegisterMsg(param *pb.AfParam, param1 *pb.ADParam, ip string, userid string) *pb.ReportRegister {
// 	if param == nil && param1 == nil {
// 		return nil
// 	}

// 	msg := &pb.ReportRegister{}

// 	if param != nil && param.AppId != "" && param.AFId != "" && param.OS != "" && param.BundleId != "" && param.MediaSource != "" && param.AFKey != "" {
// 		msg.AfInfo = &pb.AFInfo{
// 			APPID:       param.AppId,
// 			AFID:        param.AFId,
// 			OS:          param.OS,
// 			BundleID:    param.BundleId,
// 			RefGameID:   param.RefGameId,
// 			RefPkgName:  param.RefPkgName,
// 			MediaSource: param.MediaSource,
// 			AFKey:       param.AFKey,
// 		}
// 	}

// 	if param1 != nil && param1.AdAdid != "" && param1.AdKey != "" && param1.AdS2SCode != "" && param1.AdEventCode1 != "" && param1.AdEventCode2 != "" && param1.AdEventCode3 != "" && param1.AdEventCode4 != "" {
// 		msg.AdInfo = &pb.ADInfo{
// 			ADID:         param1.AdAdid,
// 			ADKey:        param1.AdKey,
// 			ADS2SCode:    param1.AdS2SCode,
// 			ADEventCode1: param1.AdEventCode1,
// 			ADEventCode2: param1.AdEventCode2,
// 			ADEventCode3: param1.AdEventCode3,
// 			ADEventCode4: param1.AdEventCode4,
// 			UserAgent:    param1.UserAgent,
// 			Ip:           ip,
// 			BundleID:     param1.BundleId,
// 			Userid:       userid,
// 		}
// 	}

// 	return msg
// }

func ReportLoginMsg(user *data.User) *pb.ReportLogin {
	msg := &pb.ReportLogin{}

	// if param != nil && param.AppId != "" && param.AFId != "" && param.OS != "" && param.BundleId != "" && param.MediaSource != "" && param.AFKey != "" {
	// 	msg.AfInfo = &pb.AFInfo{
	// 		APPID:       param.AppId,
	// 		AFID:        param.AFId,
	// 		OS:          param.OS,
	// 		BundleID:    param.BundleId,
	// 		RefGameID:   param.RefGameId,
	// 		RefPkgName:  param.RefPkgName,
	// 		MediaSource: param.MediaSource,
	// 		AFKey:       param.AFKey,
	// 	}
	// }

	if user.AD_Key != "" && user.AD_S2S_Code != "" && user.AD_Event_Code1 != "" && user.AD_Event_Code2 != "" && user.AD_Event_Code3 != "" && user.AD_Event_Code4 != "" {
		msg.AdInfo = &pb.ADInfo{
			ADID:         user.AD_ADID,
			ADKey:        user.AD_Key,
			ADS2SCode:    user.AD_S2S_Code,
			ADEventCode1: user.AD_Event_Code1,
			ADEventCode2: user.AD_Event_Code2,
			ADEventCode3: user.AD_Event_Code3,
			ADEventCode4: user.AD_Event_Code4,
			UserAgent:    user.AD_User_Agent,
			Ip:           user.RegistIP,
			BundleID:     user.AD_BundleId,
			Userid:       user.Userid,
		}
	}

	return msg
}

// func ReportLoginMsg(param *pb.AfParam, param1 *pb.ADParam, ip string, userid string) *pb.ReportLogin {
// 	if param == nil && param1 == nil {
// 		return nil
// 	}
// 	msg := &pb.ReportLogin{}

// 	if param != nil && param.AppId != "" && param.AFId != "" && param.OS != "" && param.BundleId != "" && param.MediaSource != "" && param.AFKey != "" {
// 		msg.AfInfo = &pb.AFInfo{
// 			APPID:       param.AppId,
// 			AFID:        param.AFId,
// 			OS:          param.OS,
// 			BundleID:    param.BundleId,
// 			RefGameID:   param.RefGameId,
// 			RefPkgName:  param.RefPkgName,
// 			MediaSource: param.MediaSource,
// 			AFKey:       param.AFKey,
// 		}
// 	}

// 	if param1 != nil && param1.AdAdid != "" && param1.AdKey != "" && param1.AdS2SCode != "" && param1.AdEventCode1 != "" && param1.AdEventCode2 != "" && param1.AdEventCode3 != "" && param1.AdEventCode4 != "" {
// 		msg.AdInfo = &pb.ADInfo{
// 			ADID:         param1.AdAdid,
// 			ADKey:        param1.AdKey,
// 			ADS2SCode:    param1.AdS2SCode,
// 			ADEventCode1: param1.AdEventCode1,
// 			ADEventCode2: param1.AdEventCode2,
// 			ADEventCode3: param1.AdEventCode3,
// 			ADEventCode4: param1.AdEventCode4,
// 			UserAgent:    param1.UserAgent,
// 			Ip:           ip,
// 			BundleID:     param1.BundleId,
// 			Userid:       userid,
// 		}
// 	}

// 	return msg
// }

func ReportDepositMsg(param *pb.AfParam, param1 *pb.ADParam, amount float64, user *data.User) *pb.ReportDeposit {
	if param == nil && param1 == nil {
		return nil
	}

	msg := &pb.ReportDeposit{}

	if param != nil && param.AppId != "" && param.AFId != "" && param.OS != "" && param.BundleId != "" && param.MediaSource != "" && param.AFKey != "" {
		msg.AfInfo = &pb.AFInfo{
			APPID:       param.AppId,
			AFID:        param.AFId,
			OS:          param.OS,
			BundleID:    param.BundleId,
			RefGameID:   param.RefGameId,
			RefPkgName:  param.RefPkgName,
			MediaSource: param.MediaSource,
			AFKey:       param.AFKey,
		}
	}

	if param1 != nil && param1.AdKey != "" && param1.AdS2SCode != "" && param1.AdEventCode1 != "" && param1.AdEventCode2 != "" && param1.AdEventCode3 != "" && param1.AdEventCode4 != "" {
		msg.AdInfo = &pb.ADInfo{
			ADID:         param1.AdAdid,
			ADKey:        param1.AdKey,
			ADS2SCode:    param1.AdS2SCode,
			ADEventCode1: param1.AdEventCode1,
			ADEventCode2: param1.AdEventCode2,
			ADEventCode3: param1.AdEventCode3,
			ADEventCode4: param1.AdEventCode4,
			UserAgent:    param1.UserAgent,
			Ip:           user.RegistIP,
			BundleID:     param1.BundleId,
			Userid:       user.Userid,
		}
	}

	if user.FB_Fbc != "" {
		msg.FbInfo = new(pb.FBInfo)
	}

	msg.Amount = amount
	return msg
}
