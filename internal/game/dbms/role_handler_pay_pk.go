package dbms

import "goserver/pkg/data"

func IsPk() bool {
	return country == "PK"
}

func (a *RoleActor) SavePKWithdrawInfo(user *data.User, bankName string, state int) {
	if !IsPk() || state != data.WithdrawSuccess {
		return
	}

	if user.PkWithdrawInfo.WithdrawTimes == nil {
		user.PkWithdrawInfo.WithdrawTimes = make(map[string]int64)
	}
	user.PkWithdrawInfo.WithdrawTimes[bankName]++
	// 同步db
	user.UpdatePKWithdrawInfo()
}
