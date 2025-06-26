package login

import (
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"gopkg.in/ini.v1"
)

const (
	EfiCodeSuccess             int32 = 0
	EfiCodeFiled               int32 = 16
	EfiCodeAPIError            int32 = 19
	EfiCodeInternalServerError int32 = 999
)

var (
	efiApiUrl       = "https://swmd.6633663.com"
	efiSecretKey    = "0oCcn9"
	efiOperatorCode = "E783"
	efiProductId    = "1002"
)

func EfiInit(cfg *ini.File) {
	efiApiUrl = cfg.Section("efi").Key("efiApiUrl").Value()
	efiSecretKey = cfg.Section("efi").Key("efiSecretKey").Value()
	efiOperatorCode = cfg.Section("efi").Key("efiOperatorCode").Value()
	efiProductId = cfg.Section("efi").Key("efiProductId").Value()
}

// efi 视讯

// 请求签名
// Signature = MD5(OperatorCode + RequestTime + MethodName +SecretKey)
// MD5(ABCD + 20230315104218 + getbalance + ABCD1234)
func EfiSign(methodName, requestTime string) (sign string) {
	signStr := efiOperatorCode + requestTime + strings.ToLower(methodName) + efiSecretKey
	sign = utils.Md5(signStr)
	return
}

func currentTime() (now time.Time, requestTime string) {
	now = time.Now()
	format := "20060102150405"
	requestTime = now.Format(format)
	return
}

func efiRequest(url string, requestBody map[string]string) (resp string, err error) {
	// 创建一个 Resty 客户端
	client := resty.New()

	if proxyUrl != "" {
		// 设置代理服务器地址和端口
		client.SetProxy(proxyUrl)
	}

	// 设置请求头，指定 Content-Type 为 application/json
	client.SetHeader("Content-Type", "application/json")

	// 发起 POST 请求
	r, err := client.R().SetBody(requestBody).Post(efiApiUrl + url)
	// str, _ := json.Marshal(requestBody)
	// glog.Infof("url=%s, body=%s", efiApiUrl+url, string(str))
	if r.StatusCode() != 200 {
		err = fmt.Errorf("%s", r.Status())
	}
	if err == nil {
		resp = r.String()
	}
	return
}

func EfiGetGameList(userid, username, clientIp string) (resp string, err error) {
	_, requestTime := currentTime()
	sign := EfiSign("GetGameList", requestTime)
	params := map[string]string{
		"OperatorCode": efiOperatorCode,
		"MemberName":   userid,
		"DisplayName":  username,
		"ProductID":    efiProductId,
		"GameType":     "2",
		"LanguageCode": "1",
		"Platform":     "1",
		"IPAddress":    clientIp,
		"Sign":         sign,
		"RequestTime":  requestTime,
	}
	return efiRequest("/Seamless/GetGameList", params)
}

func EfiLaunchGame(userid, username, clientIp, gameId string) (resp string, err error) {
	_, requestTime := currentTime()
	sign := EfiSign("LaunchGame", requestTime)
	params := map[string]string{
		"OperatorCode": efiOperatorCode,
		"MemberName":   userid,
		"DisplayName":  username,
		"GameID":       gameId, // GetGameList gameId
		"ProductID":    efiProductId,
		"GameType":     "2",
		"LanguageCode": "1",
		"Platform":     "1",
		"IPAddress":    clientIp,
		"Sign":         sign,
		"RequestTime":  requestTime,
	}
	return efiRequest("/Seamless/LaunchGame", params)
}

// ================================ efi callback handler =======================================
func efiParseRequest(ctx *fasthttp.RequestCtx, method string) (req *EfiRequest, req2 *pb.EfiTransactionReq, err error) {
	req = &EfiRequest{}
	body := ctx.PostBody()
	clientIP := getIP(ctx)
	glog.Infof("efi callback: ip=%s, method=%s,  body=%s", clientIP, method, string(body))

	err = jsoniter.Unmarshal(body, req)
	if err != nil {
		glog.Errorf("json unmarsals fail", err)
		rsp := fmt.Sprintf("{\"ErrorCode\":%d,\"ErrorMessage\":\"%s\"}", EfiCodeInternalServerError, err.Error())
		efiWriteResponse(ctx, string(rsp))
		return
	}
	req2 = &pb.EfiTransactionReq{
		MemberName:   req.MemberName,
		OperatorCode: req.OperatorCode,
		ProductID:    req.ProductID,
		MessageID:    req.MessageID,
		RequestTime:  req.RequestTime,
	}
	if len(req.Transactions) > 0 {
		for _, trans := range req.Transactions {
			trans2 := &pb.EfiTransaction{
				MemberID:          trans.MemberID,
				OperatorID:        trans.OperatorID,
				ProductID:         trans.ProductID,
				ProviderID:        trans.ProviderID,
				ProviderLineID:    trans.ProviderLineID,
				WagerID:           trans.WagerID,
				CurrencyID:        trans.CurrencyID,
				GameType:          trans.GameType,
				GameID:            trans.GameID,
				GameRoundID:       trans.GameRoundID,
				ValidBetAmount:    trans.ValidBetAmount,
				BetAmount:         trans.BetAmount,
				TransactionAmount: trans.TransactionAmount,
				TransactionID:     trans.TransactionID,
				PayoutAmount:      trans.PayoutAmount,
				PayoutDetail:      trans.PayoutDetail,
				CommissionAmount:  trans.CommissionAmount,
				JackpotAmount:     trans.JackpotAmount,
				JPBet:             trans.JPBet,
				Status:            trans.Status,
				SettlementDate:    trans.SettlementDate,
				CreatedOn:         trans.CreatedOn,
				ModifiedOn:        trans.ModifiedOn,
			}
			req2.Transactions = append(req2.Transactions, trans2)
		}
	}
	return
}

func efiValidateSign(ctx *fasthttp.RequestCtx, sign, method, requestTime string) error {
	sign2 := EfiSign(method, requestTime)
	if sign2 != sign {
		glog.Errorf("efiValidateSign sign error: %s", method)
		rsp := fmt.Sprintf("{\"ErrorCode\":%d,\"ErrorMessage\":\"%s\"}", EfiCodeFiled, "sign error")
		efiWriteResponse(ctx, string(rsp))
		return errors.New("sign error")
	}
	return nil
}

func efiWriteResponse(ctx *fasthttp.RequestCtx, rsp interface{}) {
	ctx.Response.Header.Set("Content-Type", "application/json")

	var response string
	if rspStr, ok := rsp.(string); ok {
		response = rspStr
	} else {
		if r, ok := rsp.(*pb.EfiTransactionRsp); ok && r.ErrCode != 0 {
			glog.Errorf("efi res error: %d, %s", r.ErrCode, r.Err)
		}
		rspBytes, err := jsoniter.Marshal(rsp)
		if err != nil {
			glog.Error("efiResponse marshal response error: %v, %v", err, rsp)
			return
		}
		response = string(rspBytes)
	}
	fmt.Fprintf(ctx, "%s", response)
}

// {"MemberName":"100001","OperatorCode":"E783","ProductID":1002,"MessageID":"240429031900824_100001","Sign":"55d26680b8e16d414809c0ea48b9233e","RequestTime":"20240429031900"}
type EfiRequest struct {
	MemberName   string           `json:"MemberName"`   // 运营商中玩家的唯一标识符
	OperatorCode string           `json:"OperatorCode"` // Seamless中运营商的唯一标识符,BO登录用户名
	ProductID    int32            `json:"ProductID"`    // Seamless中产品的唯一标识符
	MessageID    string           `json:"MessageID"`    // 当前API请求的唯一标识符
	Sign         string           `json:"Sign"`         // 请求的签名
	RequestTime  string           `json:"RequestTime"`  // 请求日期时间 日期时间格式为yyyyMMddHHmmss
	Transactions []EfiTransaction `json:"Transactions"` // 交易对象列表
}

type EfiTransaction struct {
	MemberID          int64   `json:"MemberID"`          // Seamless中玩家的唯一标识符
	OperatorID        int64   `json:"OperatorID"`        // Seamless中运营商的唯一标识符,有时它被称为AgentID
	ProductID         int64   `json:"ProductID"`         // Seamless中产品的唯一标识符
	ProviderID        int32   `json:"ProviderID"`        // Seamless 中供应商的唯一标识符。
	ProviderLineID    int32   `json:"ProviderLineID"`    // Seamless 中对产品线配置的唯一标识符。
	WagerID           int64   `json:"WagerID"`           // Seamless 投注记录的唯一标识符
	CurrencyID        int32   `json:"CurrencyID"`        // Seamless中货币的唯一标识符 INR=16
	GameType          int32   `json:"GameType"`          // 交易的游戏类型
	GameID            string  `json:"GameID"`            // 供应商游戏代码
	GameRoundID       string  `json:"GameRoundID"`       // 供应商游戏轮次ID
	ValidBetAmount    float64 `json:"ValidBetAmount"`    // 在扣除营业额赢利后的投注金额
	BetAmount         float64 `json:"BetAmount"`         // 整个下注金额，未扣除营业额中奖金额
	TransactionAmount float64 `json:"TransactionAmount"` // 需要对玩家钱包进行更改的金额, 正值表示增加玩家钱包金额，负值表示减少玩家钱包金额
	TransactionID     string  `json:"TransactionID"`     // 当前交易的唯一标识符,请返回错误代码1003以表示检测到重复交易。
	PayoutAmount      float64 `json:"PayoutAmount"`      // 玩家的赢取金额，如果玩家输了，这个值可以为0。
	PayoutDetail      string  `json:"PayoutDetail"`      // 由供应商发送的有关玩家赢取的详细信息。
	CommissionAmount  float64 `json:"CommissionAmount"`  // 佣金金额
	JackpotAmount     float64 `json:"JackpotAmount"`     // 奖池金额
	SettlementDate    string  `json:"SettlementDate"`    // 最终确认投注记录的日期,当投注结束时，玩家要么输了要么赢了。
	JPBet             float64 `json:"JPBet"`             // 供应商每次下注的奖池贡献金额,另一个术语是渐进式累计奖池
	Status            int32   `json:"Status"`            // 当前交易的状态 100 Pending, 101 Settle, 102 Void
	CreatedOn         string  `json:"CreatedOn"`         // 交易日期
	ModifiedOn        string  `json:"ModifiedOn"`        // 修改交易的日期
}

type EfiResponse struct {
	ErrorCode     int32   `json:"ErrorCode"`
	ErrorMessage  string  `json:"ErrorMessage"`
	Balance       float64 `json:"Balance"`
	BeforeBalance float64 `json:"BeforeBalance"`
}

// 获取游戏列表
func efiGameList(ctx *fasthttp.RequestCtx) {
	// 玩家类型
	userType := 0

	userToken := string(ctx.QueryArgs().Peek("userToken"))
	if userToken == "" {
		fmt.Fprintf(ctx, "%s", "userToken can not be empty")
		return
	}

	uid, err := handler.Valid(userToken)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "token invalid")
		return
	}

	msg := &pb.GetUserInfoReq{Uid: uid}
	res, err := callNode(msg)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "get user info error")
		return
	}

	var recharge int64
	var nickName string
	if rsp, ok := res.(*pb.GetUserInfoRsp); ok {
		if rsp.Err == "" {
			recharge = rsp.Recharge
			userType = int(rsp.UserType)
			nickName = rsp.NickName
		} else {
			fmt.Fprintf(ctx, "%s", rsp.Err)
			return
		}
	} else {
		fmt.Fprintf(ctx, "%s", "call node error")
		return
	}

	if recharge < 500*100 && userType == 0 {
		fmt.Fprintf(ctx, "%s", "500")
		return
	}

	// B类充200
	if recharge < 200*100 && userType == 1 {
		fmt.Fprintf(ctx, "%s", "200")
		return
	}

	clientIP := getIP(ctx)
	resp, err := EfiGetGameList(uid, nickName, clientIP)
	if err != nil {
		glog.Errorf("efiGameList error", err)
		fmt.Fprint(ctx, err.Error())
		return
	}
	fmt.Fprintf(ctx, "%s", resp)
}

// 启动游戏
func efiLaunchGame(ctx *fasthttp.RequestCtx) {
	// 玩家类型
	userType := 0

	//参数获取
	gameId := string(ctx.QueryArgs().Peek("gameId"))
	if gameId == "" {
		fmt.Fprintf(ctx, "%s", "gameId can not be empty")
		return
	}

	userToken := string(ctx.QueryArgs().Peek("userToken"))
	if userToken == "" {
		fmt.Fprintf(ctx, "%s", "userToken can not be empty")
		return
	}

	uid, err := handler.Valid(userToken)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "token invalid")
		return
	}

	msg := &pb.GetUserInfoReq{Uid: uid}
	res, err := callNode(msg)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "get user info error")
		return
	}

	var recharge int64
	var nickName string
	if rsp, ok := res.(*pb.GetUserInfoRsp); ok {
		if rsp.Err == "" {
			recharge = rsp.Recharge
			userType = int(rsp.UserType)
			nickName = rsp.NickName
		} else {
			fmt.Fprintf(ctx, "%s", rsp.Err)
			return
		}
	} else {
		fmt.Fprintf(ctx, "%s", "call node error")
		return
	}

	if recharge < 500*100 && userType == 0 {
		fmt.Fprintf(ctx, "%s", "500")
		return
	}

	// B类充200
	if recharge < 200*100 && userType == 1 {
		fmt.Fprintf(ctx, "%s", "200")
		return
	}

	clientIP := getIP(ctx)
	resp, err := EfiLaunchGame(uid, nickName, clientIP, gameId) // rkfaxspyyoeqae3g
	if err != nil {
		fmt.Fprintf(ctx, "%s", err.Error())
		return
	}
	fmt.Fprintf(ctx, "%s", resp)
}

// 2.1 获得余额
func efiGetBalance(ctx *fasthttp.RequestCtx) {
	rsp := &EfiResponse{}
	arg, _, err := efiParseRequest(ctx, "GetBalance")
	if err != nil {
		return
	}
	if err := efiValidateSign(ctx, arg.Sign, "GetBalance", arg.RequestTime); err != nil {
		return
	}

	// get balance
	msg := &pb.GetUserInfoReq{Uid: arg.MemberName}
	res, err := callNode(msg)
	if err != nil {
		rsp.ErrorCode = EfiCodeInternalServerError
		rsp.ErrorMessage = "request error"
		efiWriteResponse(ctx, rsp)
		return
	}
	user, ok := res.(*pb.GetUserInfoRsp)
	if !ok || user.Err != "" {
		rsp.ErrorCode = EfiCodeInternalServerError
		rsp.ErrorMessage = "data error"
		if user.Err != "" {
			rsp.ErrorMessage = user.Err
		}
		efiWriteResponse(ctx, rsp)
		return
	}
	balance := float64(user.Balance) / 100.0
	rsp.Balance = balance
	glog.Debugf("efi get balance: %f", balance)

	efiWriteResponse(ctx, rsp)
}

func efiTrasaction(ctx *fasthttp.RequestCtx, method string, transType pb.FfiTransactionType) {
	rsp := &EfiResponse{}
	arg, req, err := efiParseRequest(ctx, method)
	if err != nil {
		return
	}
	if err := efiValidateSign(ctx, arg.Sign, method, arg.RequestTime); err != nil {
		return
	}

	req.TransType = transType
	res, err := callNode(req)
	if err != nil {
		rsp.ErrorCode = EfiCodeInternalServerError
		rsp.ErrorMessage = "request error"
		efiWriteResponse(ctx, rsp)
		return
	}
	r, ok := res.(*pb.EfiTransactionRsp)
	if !ok {
		rsp.ErrorCode = EfiCodeInternalServerError
		rsp.ErrorMessage = "data error"
		efiWriteResponse(ctx, rsp)
		return
	}
	if r.ErrCode != 0 || r.Err != "" {
		rsp.ErrorCode = EfiCodeInternalServerError
		if r.ErrCode != 0 {
			rsp.ErrorCode = int32(r.ErrCode)
		}
		rsp.ErrorMessage = r.Err
		efiWriteResponse(ctx, rsp)
		return
	}

	rsp.ErrorCode = 0
	rsp.ErrorMessage = "Success"
	rsp.Balance = r.Balance
	rsp.BeforeBalance = r.BeforeBalance
	efiWriteResponse(ctx, rsp)
}

// 2.2 下注
func efiPlaceBet(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "PlaceBet", pb.EfiPlaceBet)
}

// 2.3 游戏结果
func efiGameResult(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "GameResult", pb.EfiGameResult)
}

// 2.4 回滚
func efiRollback(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "Rollback", pb.EfiRollback)
}

// 2.5 取消下注
func efiCancelBet(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "CancelBet", pb.EfiCancelBet)
}

// 2.6 奖金/红利
func efiBonus(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "Bonus", pb.EfiBonus)
}

// 2.7 奖池
func efiJackpot(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "Jackpot", pb.EfiJackpot)
}

// 2.8 手机登录
func efiMobileLogin(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	glog.Infof("efiMobileLogin: %s", string(body))
}

// 2.9 买入
func efiBuyIn(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "BuyIn", pb.EfiBuyIn)
}

// 2.10 买断
func efiBuyOut(ctx *fasthttp.RequestCtx) {
	efiTrasaction(ctx, "BuyOut", pb.EfiBuyOut)
}

// 2.11 下注推送
func efiPushBet(ctx *fasthttp.RequestCtx) {
	body := ctx.PostBody()
	glog.Infof("efiPushBet: %s", string(body))
}
