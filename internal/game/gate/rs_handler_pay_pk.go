package gate

import (
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/ck"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/table"
	"goserver/pkg/utils"
	"sort"
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/mgo.v2/bson"
)

func IsPk() bool {
	return country == "PK"
}

// 巴基斯坦提现
func (rs *RoleActor) PakWithDrawReq(ctx actor.Context) {
	msg := ctx.Message()
	arg := msg.(*pb.PakWithDrawReq)
	glog.Debugf("PakWithDrawReq %#v", arg)
	rs.pakWithDrawOrder(arg)
}

// 巴基斯坦提现
func (rs *RoleActor) pakWithDrawOrder(arg *pb.PakWithDrawReq) {
	user := rs.User
	rsp := new(pb.PakWithDrawRsp)
	// 绑没绑定手机
	if user.Phone == "" {
		rsp.Error = pb.NoBindPhone
		rs.Send(rsp)
		return
	}
	// 能不能下单
	if ok, code, param := handler.CanWithDraw(user, arg.Amount); !ok {
		glog.Errorf("user %s can't create withdraworder", rs.Userid)
		rsp.Error = code
		rsp.Param = param
		rs.Send(rsp)
		return
	}

	// 今日提现次数金额限制二次确认
	now := time.Now().In(location)
	stime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	etime := stime.AddDate(0, 0, 1)
	var withdraw_data = make(map[string]any)
	err := ck.Select(&withdraw_data, `
		select count(*) withdraw_times, SUM(amount) withdraw_amounts
		from game.col_withdraw_record final 
		where ctime >= ? and ctime < ? and userid = ? and order_status not in (3, 12)
	`, stime, etime, user.Userid)
	if err != nil {
		glog.Errorf("user %s withdraw order select error: ", err)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	} else {
		withdraw_times := utils.ToInt64(withdraw_data["withdraw_times"])
		withdraw_amounts := utils.ToInt64(withdraw_data["withdraw_amounts"])

		var errCode pb.ErrCode
		vip := table.GetTables().VipTable.Get(int32(user.Vip.Lv))
		if withdraw_times >= int64(vip.WithdrawTimes) && vip.WithdrawTimes != -1 {
			// 提现次数不足
			errCode = pb.WithdrawCountUnenough
		}
		if (int64(arg.Amount) > vip.WithdrawAmounts ||
			withdraw_amounts >= vip.WithdrawAmounts ||
			int64(arg.Amount)+withdraw_amounts > vip.WithdrawAmounts) &&
			vip.WithdrawAmounts != -1 {
			// 提现金额不足
			errCode = pb.WithdrawAmountUnenough
		}
		if errCode != pb.OK {
			glog.Errorf("user %s can't create withdraworder second check: vip=%d, db=(%d, %d), user=(%d, %d)", rs.Userid, user.Vip.Lv,
				withdraw_times, withdraw_amounts,
				user.Vip.WithdrawCount, user.Vip.WithdrawAmount,
			)
			rsp.Error = errCode
			rs.Send(rsp)
			return
		}
	}

	// 保存银行信息(不可修改的)
	rs.savePkBank(arg)
	// 创建订单
	order, err := handler.CreatePKWithDrawOrder(rs.User, arg)
	if err != nil {
		glog.Errorf("user %s create withdraworder fail", rs.Userid)
		glog.Error(err)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	}
	// 保存订单记录
	body, err1 := jsoniter.Marshal(order)
	if err1 != nil {
		glog.Errorf("user %s create serialize fail, err:%v", rs.Userid, err1)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	}
	if !rs.saveOrder(2, 1, body, order.OrderID) {
		glog.Errorf("user %s save withdraw order fail, err:%v", rs.Userid, err1)
		rsp.Error = pb.WithdrawOrderError
		rs.Send(rsp)
		return
	}
	// 扣钱
	rs.addCurrency(-int64(order.Score+order.Commission), 0, 0, 0, 0, -int64(order.Score), 0, int32(pb.LOG_TYPE13), "提现", order.OrderID)
	// rs.sendGood(-int64(order.Score), 0, 0, -int64(order.Score), 0, 0, int32(pb.LOG_TYPE13), "提现")
	// 增加提现次数
	if user.WithDrawCountMap == nil {
		user.WithDrawCountMap = make(map[int32]int32)
	}
	count := user.WithDrawCountMap[arg.Id]
	user.WithDrawCountMap[arg.Id] = count + 1
	user.Vip.WithdrawCount++
	user.Vip.WithdrawAmount += int(arg.Amount)

	user.TpUserTodayWithdrawNum++
	user.TpUserTodayWithdrawAmount += int64(order.Score)

	rsp.Id = arg.Id
	rsp.Cash = user.Diamond
	rsp.Count = count + 1
	rsp.WithdrawCount = int32(user.Vip.WithdrawCount)
	rsp.WithdrawAmount = int32(user.Vip.WithdrawAmount)
	rs.Send(rsp)
	// 添加订单记录
	rs.addWithdrawLog(user, order)
	// 发邮件
	ntf := new(pb.FeedBackLogNtf)
	chat := data.ChatLog{
		Uid:      bson.NewObjectId().String(),
		Receiver: rs.Userid,
		Title:    "System Message",
		Content:  handler.BuildWithdrawApply(user.Nickname),
		Ctime:    utils.LocalTime(),
		Sender:   "-1",
		Name:     "System",
	}
	user.FeedBackLogMap[chat.Uid] = chat
	rs.status = true
	ntf.Logs = append(ntf.Logs, &pb.ChatLog{Uid: chat.Uid, Userid: chat.Sender, Title: chat.Title, Ctime: chat.Ctime.Unix(), Content: chat.Content})
	rs.Send(ntf)
}

func (rs *RoleActor) savePkBank(arg *pb.PakWithDrawReq) {
	for _, bank := range rs.PKBank {
		if bank.BankName == arg.BankName && bank.BankAccount == arg.BankNumber {
			return
		}
	}
	rs.PKBank = append(rs.PKBank, data.PKBankInfo{
		BankName:    arg.BankName,
		BankAccount: arg.BankNumber,
		BankHolder:  arg.BankHolder,
	})
	rs.status = true
}

func (rs *RoleActor) BuildPKWithdrawData(AppsBanks, AppsWallets map[string]struct{}) *pb.PAKWithDrawData {
	pkData := &pb.PAKWithDrawData{
		// DefaultType: 1,
	}

	if rs.PkWithdrawInfo.WithdrawTimes == nil {
		rs.PkWithdrawInfo.WithdrawTimes = make(map[string]int64)
	}

	type tmpTimes struct {
		BankName string
		Times    int64
	}

	var banksTimes, walletsTimes []tmpTimes

	for bank, times := range rs.PkWithdrawInfo.WithdrawTimes {
		if _, ok := AppsBanks[bank]; ok {
			banksTimes = append(banksTimes, tmpTimes{
				BankName: bank,
				Times:    times,
			})
			delete(AppsBanks, bank)
		}
		if _, ok := AppsWallets[bank]; ok {
			walletsTimes = append(banksTimes, tmpTimes{
				BankName: bank,
				Times:    times,
			})
			delete(AppsWallets, bank)
		}
	}

	sort.Slice(banksTimes, func(i, j int) bool {
		return banksTimes[i].Times > banksTimes[j].Times
	})
	sort.Slice(walletsTimes, func(i, j int) bool {
		return walletsTimes[i].Times > walletsTimes[j].Times
	})

	for _, bank := range banksTimes {
		pkData.Banks = append(pkData.Banks, bank.BankName)
	}
	for _, wallet := range walletsTimes {
		pkData.Wallets = append(pkData.Wallets, wallet.BankName)
	}

	for k := range AppsBanks {
		pkData.Banks = append(pkData.Banks, k)
	}

	for k := range AppsWallets {
		pkData.Wallets = append(pkData.Wallets, k)
	}
	return pkData
}

func (rs *RoleActor) SavePKWithdrawInfo(bankName string) {
	if !IsPk() {
		return
	}
	if rs.PkWithdrawInfo.WithdrawTimes == nil {
		rs.PkWithdrawInfo.WithdrawTimes = make(map[string]int64)
	}
	rs.PkWithdrawInfo.WithdrawTimes[bankName]++
	rs.status = true
}

func (rs *RoleActor) WithdrawBankInfo(arg *pb.WithdrawBankInfoReq) {
	rsp := new(pb.WithdrawBankInfoRsp)
	defer rs.Send(rsp)
	if arg.Option == 1 { // 新增
		for _, bank := range rs.PKBank {
			if bank.BankName == arg.BankName && bank.BankAccount == arg.BankNumber {
				// 已经存在了
				rsp.Error = pb.BankAlreadyExist
				return
			}
		}
		rs.PKBank = append(rs.PKBank, data.PKBankInfo{
			BankName:    arg.BankName,
			BankAccount: arg.BankNumber,
			BankHolder:  arg.BankHolder,
		})
		return
	}
	// 删除
	index := -1
	for i, bank := range rs.PKBank {
		if bank.BankName == arg.BankName && bank.BankAccount == arg.BankNumber {
			index = i
			break
		}
	}
	if index == -1 {
		rsp.Error = pb.Failed
		return
	}
	rs.PKBank = append(rs.PKBank[:index], rs.PKBank[index+1:]...)
	rs.status = true
}
