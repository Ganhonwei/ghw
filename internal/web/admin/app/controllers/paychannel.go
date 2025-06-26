package controllers

import (
	"encoding/json"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/data"
	"goserver/pkg/game/handler"
	"goserver/pkg/utils"
	"net/url"
	"strconv"
	"time"

	"github.com/astaxie/beego"
)

var (
	payChannelBalances map[string]string
)

func init() {
	// 定时查询通道余额
	if service.RunMode == "dev" {
		payChannelBalances = make(map[string]string)
	} else {
		go func() {
			if _, err := getPayChannelBalance(true); err != nil {
				beego.Error("fetch pay channel balance error:", err)
			}

			ticker := time.NewTicker(time.Hour)
			for {
				<-ticker.C
				beego.Debug("fetch pay channel balance...")
				if _, err := getPayChannelBalance(true); err != nil {
					beego.Error("fetch pay channel balance error:", err)
				}
			}
		}()
	}
}

func getPayChannelBalance(force bool) (bss map[string]string, err error) {
	if service.RunMode == "dev" {
		return payChannelBalances, nil
	}

	if !force && (payChannelBalances != nil) {
		return payChannelBalances, nil
	}

	rsp, err := service.PayRequestGet(pb.WebPayChannel, "/querybalance", url.Values{})
	if err != nil {
		return
	}
	var balances []entity.WebQueryBalence
	if err = json.Unmarshal(rsp.Body, &balances); err != nil {
		return
	}
	bss = make(map[string]string)
	for _, bs := range balances {
		bss[fmt.Sprint(bs.ChannelId)] = fmt.Sprintf("%.2f", float64(bs.Balence)/100)
	}
	payChannelBalances = bss
	return
}

// 充值渠道
func (this *PayController) PayChannel() {
	action := this.GetString("action")
	// 通道成功率
	if action == "HourPaySuccessRate" {
		rates, err := service.PayService.PayChannelHourPaySuccessRate()
		if err != nil {
			this.jsonResult(libs.RError(err.Error()))
			return
		}
		payRates := make(map[string]string)
		for id, rate := range rates {
			payRates[id] = fmt.Sprintf("%.2f%%", rate*100)
		}
		// this.jsonResult(libs.RSuccess(payRates))
		this.jsonResult(libs.RSuccess(map[string]any{
			"rates": payRates,
			"time":  time.Now().In(service.Location()).Format(utils.FORMAT_TIME),
		}))
		return
	}

	// 通道余额
	if action == "BalanceQuery" {
		force, _ := this.GetInt("force")
		bss, err := getPayChannelBalance(force == 1)
		if err != nil {
			this.JsonRFail(err.Error())
			return
		}
		this.jsonResult(libs.RSuccess(map[string]any{
			"bss":  bss,
			"time": time.Now().In(service.Location()).Format(utils.FORMAT_TIME),
		}))
		return
	}

	if action == "Aside" {
		this.JsonRSuccess(libs.R{
			"isOperation": this.auth.HasAccessPerm(this.controllerName, "paychannelopt"),
		})
		return
	}

	if action == "PayChannels" {
		_, _, sortBy, asc := this.PageArgs()
		var sort int // 0.默认,1代收降序,2.代付降序
		if sortBy == "switch" {
			if asc == 1 {
				sort = 1
			} else {
				sort = 2
			}
		}
		total, list, err := service.PayService.PayChannelList(sort)
		if err != nil {
			this.JsonRError(err.Error())
			return
		}

		this.JsonRSuccess(libs.R{
			"list":    list,
			"total":   total,
			"version": beego.AppConfig.String("versions"),
		})
		return
	}

	this.Data["pageTitle"] = "通道配置"
	this.display()
}

func payCahnnelCheck(c *entity.PayChannel) (err string) {
	version := beego.AppConfig.String("versions")
	if c.Status != 0 {
		if len(c.PayOptions) == 0 && !c.UtrRequired && version != "PK" {
			return "开启代收,请配置支持的支付方式/是否需要UTR"
		}
		if len(c.PayApps) == 0 {
			return "开启代收,请配置支持的支付APP"
		}
		if c.PayWeight <= 0 {
			return "开启代收,请设置代收权重"
		}
		if c.PayMin == 0 || c.PayMax == 0 || c.PayMin > c.PayMax {
			return "开启代付,请设置充值金额区间"
		}
	}
	if c.Wstatus != 0 {
		if c.WithdrawWeight <= 0 {
			return "开启代付,请设置代付权重"
		}
		if c.WithdrawMin == 0 || c.WithdrawMax == 0 || c.WithdrawMin > c.WithdrawMax {
			return "开启代付,请设置提现金额区间"
		}
	}
	return
}

// 充值渠道编辑
func (this *PayController) PayChannelEdit() {
	action := this.GetString("action")

	// 开启关闭渠道开关
	if action == "EditSwitch" {
		payid := this.GetString("payid")
		paytype := this.GetString("paytype") // 1代收 2代付
		open, _ := this.GetInt("open")
		if !utils.SliceIn(paytype, "1", "2") || !utils.SliceIn(open, 0, 1) {
			this.JsonRFail("请求参数有误")
			return
		}

		payChannel := service.PayService.GetPayChannelById(payid)
		if payChannel.Id == "" {
			this.JsonRFail("通道不存在")
			return
		}
		if paytype == "1" {
			payChannel.Status = open
		} else {
			payChannel.Wstatus = open
		}
		if err := payCahnnelCheck(payChannel); err != "" {
			this.JsonRFail(err)
			return
		}

		if err := service.PayService.AddOrUpdatePayChannel(payChannel); err != nil {
			this.JsonRError("操作失败:" + err.Error())
			return
		}

		// 通知pay服务器
		b := []entity.PayChannel{*payChannel}
		_, err := service.PayRequest(pb.WebPayChannel, "/syncpaychannel", b)
		if err != nil {
			this.JsonRError("同步失败:" + err.Error())
			return
		}

		service.ActionService.Add("set_pay_channel", this.auth.GetUser().UserName,
			"", utils.String(payid), utils.String(payid), "")

		this.JsonRSuccess(libs.R{})
		return
	}

	// 渠道编辑
	if action == "EditChannel" {
		arg := new(args.ChannelEditArgs)
		this.JsonBody(arg)

		payChannel := &entity.PayChannel{
			Id:              arg.Id,
			Name:            arg.Name,
			SortId:          arg.SortId,
			Status:          arg.Status,
			Wstatus:         arg.Wstatus,
			Ptype:           arg.Ptype,
			PayRate:         arg.PayRate,
			WithdrawRate:    arg.WithdrawRate,
			WithdrawFee:     arg.WithdrawFee,
			PayOptions:      arg.PayOptions,
			PayApps:         arg.PayApps,
			WithdrawBanks:   arg.WithdrawBanks,
			WithdrawWallets: arg.WithdrawWallets,
			UtrRequired:     arg.UtrRequired,
			PayWeight:       arg.PayWeight,
			PayMin:          arg.PayMin,
			PayMax:          arg.PayMax,
			WithdrawWeight:  arg.WithdrawWeight,
			WithdrawMin:     arg.WithdrawMin,
			WithdrawMax:     arg.WithdrawMax,
			Balance:         int(arg.Balance),
		}
		if err := payCahnnelCheck(payChannel); err != "" {
			this.JsonRFail(err)
			return
		}
		if err := service.PayService.AddOrUpdatePayChannel(payChannel); err != nil {
			this.JsonRError("操作失败:" + err.Error())
			return
		}

		// 通知pay服务器
		b := []entity.PayChannel{*payChannel}
		_, err := service.PayRequest(pb.WebPayChannel, "/syncpaychannel", b)
		if err != nil {
			this.JsonRError("同步失败:" + err.Error())
			return
		}

		service.ActionService.Add("set_pay_channel", this.auth.GetUser().UserName,
			"", utils.String(arg.Id), utils.String(arg.Id), "")

		this.JsonRSuccess(libs.R{})
		return
	}

	this.JsonRFail("unknown request")
}

// 充值渠道测试
func (this *PayController) PayChannelTest() {
	channelIdS := this.GetString("id")
	amountS := this.GetString("amount")
	channelId, err := strconv.Atoi(channelIdS)
	if err != nil || channelId <= 0 {
		this.JsonRFail("通道id有误")
		return
	}
	amount, err := strconv.ParseInt(amountS, 10, 64)
	if err != nil || amount <= 0 {
		this.JsonRFail("金额有误")
		return
	}
	amount *= 100

	userName := this.auth.GetUser().UserName
	// 通知pay服务器
	orderId := fmt.Sprintf("web-%d", time.Now().Unix())
	req := data.WebPayRequest{
		OrderID:    orderId,
		BusinessID: orderId,
		ChannelId:  uint32(channelId),
		Amount:     uint32(amount),
		Userid:     userName,
		RegistIp:   "171.79.149.92",
	}
	rsp := new(data.PayResponse)
	err = service.PayRequestJson(pb.WebPayChannel, "/webpayrequesthandler", req, rsp)
	if err != nil {
		this.JsonRError("请求pay节点失败:" + err.Error())
		return
	}

	// fmt.Println("test pay response", string(response.Body))
	// rsp := new(data.PayResponse)
	// if err = json.Unmarshal(response.Body, rsp); err != nil {
	// 	this.jsonResult(libs.RError(err.Error()))
	// 	return
	// }
	if rsp.Code != 200 {
		this.JsonRError(fmt.Sprintf("%d: %s", rsp.Code, rsp.Msg))
		return
	}

	// 生成订单
	ip := this.getClientIp()
	order := &entity.PayRecord{
		Userid:       userName,
		NickName:     userName,
		RealName:     userName,
		Email:        fmt.Sprintf("%s@gmail.com", userName),
		Mobile:       "9876543210",
		OrderStatus:  data.Tradeing,
		Ctime:        service.NowTime(),
		OrderAddress: ip,
		PackageId:    "pcdefault",
		Gtype:        0,
		UserType:     0,
		PayWay:       0,
		Amount:       uint32(amount),
		Score:        uint32(amount),
		ShopName:     fmt.Sprintf("后台测试单:%d", amount/100),
		ShopType:     data.WEB_TEST_SHOP,
		OtherPresent: "0",
		ChannelId:    req.ChannelId,
	}
	order.ReportId = orderId
	order.OrderID = fmt.Sprint(handler.GenerateOrderId(uint32(order.Ctime.Year())))
	order.OutTradeStatus = utils.String(rsp.Code)
	order.OutTradeNo = rsp.OrderID

	if err := service.Pays.Insert(order); err != nil {
		beego.Error("后台测试订单保存失败", err)
	}

	this.JsonRSuccess(rsp)
}
