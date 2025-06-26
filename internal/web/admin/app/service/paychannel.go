package service

import (
	"errors"
	"fmt"
	"goserver/internal/web/admin/app/entity"
	"goserver/pkg/data/ck"
	"goserver/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

// 后台-通道配置
func (this *payService) PayChannelList(sortType int) (total int64, list []*entity.PayChannel, err error) {
	m := bson.M{}

	var dbChannels []*entity.PayChannel
	err = PayChannels.
		Find(m).
		Sort("sort_id"). // -sort_id
		All(&dbChannels)
	if err != nil {
		return
	}

	total = int64(len(dbChannels))

	channelPayRates, err := this.PayChannelHourPaySuccessRate()
	if err != nil {
		return
	}

	// 3021 IcePay更改名称为CloudPay
	version := beego.AppConfig.String("versions")
	channels := make([]*entity.PayChannel, 0)
	originChannels := []*entity.PayChannel{
		{Id: "3010", Name: "XDPAY", SortId: 1},
		{Id: "3011", Name: "MLPAY", SortId: 2},
		{Id: "3012", Name: "1916PAY", SortId: 3},
		{Id: "3014", Name: "KingPay", SortId: 4},
		{Id: "3015", Name: "XFPAY", SortId: 5},
		{Id: "3016", Name: "SailsPay", SortId: 6},
		{Id: "3017", Name: "FLYPAY", SortId: 7},
		{Id: "3018", Name: "UwinPay", SortId: 8},
		{Id: "3013", Name: "LetsPay", SortId: 9},
		{Id: "3020", Name: "RamaPay", SortId: 10},
		{Id: "3021", Name: "CloudPay", SortId: 11},
		{Id: "3022", Name: "WePay", SortId: 12},
		{Id: "3023", Name: "OePay", SortId: 13},
		{Id: "3024", Name: "9sPay", SortId: 14},
		{Id: "3025", Name: "BLIZZARDPAY", SortId: 14},
		{Id: "3026", Name: "MATEGOPAY", SortId: 15},
		{Id: "3027", Name: "UNIVERSALPAY", SortId: 15},
		{Id: "3028", Name: "KKPLUSPAY", SortId: 15},
		{Id: "3029", Name: "DRAGONPAY", SortId: 15},
		{Id: "3030", Name: "AIPAY", SortId: 15},
		{Id: "3031", Name: "GAMEPAY", SortId: 15},
		{Id: "3032", Name: "GENTLEPAY", SortId: 15},
		{Id: "3033", Name: "66PAY", SortId: 15},
		{Id: "3034", Name: "AI2PAY", SortId: 15},
		{Id: "3035", Name: "COWPAY", SortId: 15},
		{Id: "3036", Name: "LETSPAYFAST", SortId: 15},
		{Id: "3037", Name: "PAYWOOK", SortId: 15},
		{Id: "3038", Name: "MEPAY", SortId: 15},
		{Id: "3039", Name: "DDPAY", SortId: 15},
		{Id: "3040", Name: "ATPAY", SortId: 15},
		{Id: "3041", Name: "NETPAY", SortId: 15},
		{Id: "3042", Name: "WYPPAY", SortId: 15},
		{Id: "3043", Name: "YUNPAY", SortId: 15},
		{Id: "3044", Name: "CXPAY", SortId: 15},
		{Id: "3045", Name: "CSMPAY", SortId: 15},
		{Id: "3046", Name: "WEPAY2", SortId: 15},
		{Id: "3047", Name: "LEOPAY", SortId: 15},
		{Id: "3048", Name: "UPAY", SortId: 15},
		{Id: "3049", Name: "EOCPAY", SortId: 15},
	}
	mjlaChannels := []*entity.PayChannel{
		{Id: "5001", Name: "MJL-Universe", SortId: 15},
		// {Id: "5002", Name: "USDT", SortId: 16},
		{Id: "5003", Name: "MJL-JY", SortId: 17},
		{Id: "5004", Name: "MJL-Transafe", SortId: 18},
		{Id: "5005", Name: "MJL-MMPay", SortId: 19},
		{Id: "5006", Name: "MJL-SafePay", SortId: 20},
		{Id: "5007", Name: "MJL-99Pay", SortId: 21},
		{Id: "5008", Name: "MJL-GoPay", SortId: 22},
		{Id: "5009", Name: "MJL-SHPay", SortId: 23},
		{Id: "5010", Name: "MJL-BLIZZARDPAY", SortId: 24},
		{Id: "5011", Name: "MJL-TTT", SortId: 25},
		{Id: "5012", Name: "MJL-AI2PAY", SortId: 26},
		{Id: "5013", Name: "MJL-66PAY", SortId: 27},
		{Id: "5014", Name: "MJL-EOCPAY", SortId: 28},
	}
	pakChannels := []*entity.PayChannel{
		{Id: "6001", Name: "BJS-AI2PAY", SortId: 27},
		{Id: "6002", Name: "BJS-PAKPAY", SortId: 28},
		{Id: "6003", Name: "BJS-EOCPAY", SortId: 29},
	}
	switch version {
	case "origin":
		channels = originChannels
	case "mengjiala":
		channels = mjlaChannels
	case "PK":
		channels = pakChannels
	default:
		channels = append(channels, originChannels...)
		channels = append(channels, mjlaChannels...)
		channels = append(channels, pakChannels...)
	}

	channelsMap := make(map[string]*entity.PayChannel)
	for _, c := range dbChannels {
		channelsMap[c.Id] = c
	}

	for _, c := range channels {
		cm, ok := channelsMap[c.Id]
		if !ok {
			list = append(list, c)
			continue
		}

		c = cm
		// 近1小时代收成功率
		if payRate, ok := channelPayRates[c.Id]; ok {
			c.FPaySuccessHourRate = fmt.Sprintf("%.2f%%", payRate*100)
			c.PaySuccessHourRate = payRate
		} else {
			c.FPaySuccessHourRate = "--"
		}
		payChannelMapping(c)
		list = append(list, c)
	}
	switch sortType {
	case 1: // 代收降序
		sort.Slice(list, func(i, j int) bool {
			a := list[i].Status == list[j].Status
			b := list[i].Wstatus == list[j].Wstatus
			if a && list[i].Status == 1 {
				return list[i].Id < list[j].Id
			}
			return list[i].Status > list[j].Status ||
				(a && list[i].Wstatus > list[j].Wstatus) ||
				(a && b && list[i].Id < list[j].Id)
		})
	case 2: // 代付降序
		sort.Slice(list, func(i, j int) bool {
			a := list[i].Wstatus == list[j].Wstatus
			b := list[i].Status == list[j].Status
			if a && list[i].Wstatus == 1 {
				return list[i].Id < list[j].Id
			}
			return list[i].Wstatus > list[j].Wstatus ||
				(a && list[i].Status > list[j].Status) ||
				(a && b && list[i].Id < list[j].Id)
		})
	default: // 通道id排序
		sort.Slice(list, func(i, j int) bool {
			return list[i].Id < list[j].Id
		})
	}
	return
}

func payChannelMapping(v *entity.PayChannel) {
	// 0.未知 1.纯原生 2.纯唤醒 3.原唤混
	// 通道类型
	if ptype, ok := utils.CaseWhen3(v.Ptype, 1, "纯原生", 2, "纯唤醒", 3, "原唤混"); ok {
		v.FPtype = ptype
	} else {
		v.FPtype = "未配置"
	}

	v.FPayRate = fmt.Sprintf("%.2f%%", float64(v.PayRate))                          // 代收费率
	v.FWithdrawRate = fmt.Sprintf("%.2f%%", float64(v.WithdrawRate))                // 代付费率
	v.FWithdrawFee = fmt.Sprintf("%.2f", ComputeFloat(float64(v.WithdrawFee), 100)) // 代付单笔额外费用
	// 支持的支付方式
	var payOptions, payOptionIds []string
	for _, option := range v.PayOptions {
		payOptionIds = append(payOptionIds, strconv.Itoa(int(option)))
		switch option {
		case 1:
			payOptions = append(payOptions, "DirectLaunchTheApp")
		case 2:
			payOptions = append(payOptions, "QRcode")
		case 3:
			payOptions = append(payOptions, "FillInUTR")
		}
	}
	v.FPayOptions = strings.Join(payOptions, ",")
	v.FPayOptionIds = strings.Join(payOptionIds, ",")
	// 支持的支付app
	var payApps, payAppIds []string
	for _, app := range v.PayApps {
		payAppIds = append(payAppIds, strconv.Itoa(int(app)))
		switch app {
		case 1:
			payApps = append(payApps, "Paytm")
		case 2:
			payApps = append(payApps, "PhonePe")
		case 3:
			payApps = append(payApps, "Mobikwik")
		case 4:
			payApps = append(payApps, "BHIM")
		case 5:
			payApps = append(payApps, "GooglePay")
		case 6:
			payApps = append(payApps, "Other UPI")
		case 101:
			payApps = append(payApps, "Nagad")
		case 102:
			payApps = append(payApps, "bkash")
		case 1001:
			payApps = append(payApps, "EASYPAISA")
		case 1002:
			payApps = append(payApps, "JAZZCASH")
		}
	}
	v.FPayApps = strings.Join(payApps, ",")
	v.FPayAppIds = strings.Join(payAppIds, ",")
	v.FUtrRequired = utils.CaseElse(v.UtrRequired, "是", "否")                                          // 是否需要UTR
	v.FPayRange = fmt.Sprintf("%.2f-%.2f", Chip2Float(v.PayMin), Chip2Float(v.PayMax))                // 充值金额区间
	v.FWithdrawRange = fmt.Sprintf("%.2f-%.2f", Chip2Float(v.WithdrawMin), Chip2Float(v.WithdrawMax)) // 提现金额区间
	v.FWithdrawBanks = strings.Join(v.WithdrawBanks, ",")
	v.FWithdrawWallets = strings.Join(v.WithdrawWallets, ",")
	v.FPayMin = fmt.Sprintf("%.2f", Chip2Float(v.PayMin))
	v.FPayMax = fmt.Sprintf("%.2f", Chip2Float(v.PayMax))
	v.FWithdrawMin = fmt.Sprintf("%.2f", Chip2Float(v.WithdrawMin))
	v.FWithdrawMax = fmt.Sprintf("%.2f", Chip2Float(v.WithdrawMax))
	// 通道余额（刷新）
	// v.FBalance = fmt.Sprintf("%.2f", Chip2Float(int64(v.Balance)))
	v.FBalance = "--"
	c, _ := ConvertToIndiaTime(v.Ctime.Unix())
	v.Ctime = c
}

// 计算近1小时代收成功率（刷新）
func (this *payService) PayChannelHourPaySuccessRate() (channelPayRates map[string]float64, err error) {
	channelPayRates = make(map[string]float64)

	now := time.Now().In(location)
	start := now.Add(-time.Hour)
	var results []map[string]any
	err = ck.Select(&results, `
		SELECT channel_id, COUNT(*) pay_times, SUM(CASE order_status WHEN 4 THEN 1 ELSE 0 END) pay_success_times
		FROM game.col_trade_record FINAL
		WHERE ctime BETWEEN ? AND ?
		GROUP BY channel_id
	`, start, now)
	if err != nil {
		return
	}

	for _, r := range results {
		channel_id := fmt.Sprint(r["channel_id"])
		pay_times := utils.ToFloat64(r["pay_times"])
		pay_success_times := utils.ToFloat64(r["pay_success_times"])
		channelPayRates[channel_id] = utils.CaseElse(pay_times == 0, 0, pay_success_times/pay_times)
	}
	return
}

func (this *payService) GetPayChannelById(id string) (c *entity.PayChannel) {
	c = new(entity.PayChannel)
	GetByQ(PayChannels, bson.M{"_id": id}, c)
	return
}

// 添加或修改支付渠道
func (this *payService) AddOrUpdatePayChannel(res *entity.PayChannel) error {
	info := new(entity.PayChannel)
	GetByQ(PayChannels, bson.M{"_id": bson.M{"$eq": res.Id}}, info)
	if info.Id == "" {
		res.Ctime = bson.Now()
		if !Insert(PayChannels, res) {
			return errors.New("写入失败:" + res.Id)
		}
		return nil
	} else {
		m := bson.M{
			"name":             res.Name,
			"ptype":            res.Ptype,
			"status":           res.Status,
			"wstatus":          res.Wstatus,
			"pay_rate":         res.PayRate,
			"withdraw_rate":    res.WithdrawRate,
			"withdraw_fee":     res.WithdrawFee,
			"balance":          res.Balance,
			"sort_id":          res.SortId,
			"ctime":            bson.Now(),
			"pay_options":      res.PayOptions,
			"pay_apps":         res.PayApps,
			"utr_required":     res.UtrRequired,
			"pay_weight":       res.PayWeight,
			"pay_min":          res.PayMin,
			"pay_max":          res.PayMax,
			"withdraw_weight":  res.WithdrawWeight,
			"withdraw_min":     res.WithdrawMin,
			"withdraw_max":     res.WithdrawMax,
			"withdraw_banks":   res.WithdrawBanks,
			"withdraw_wallets": res.WithdrawWallets,
		}
		if Update(PayChannels, bson.M{"_id": res.Id}, bson.M{"$set": m}) {
			return nil
		}
		return errors.New("更新失败")
	}
}
