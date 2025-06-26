package login

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/game/config"
	"goserver/pkg/game/handler"
	"goserver/pkg/glog"
	"goserver/pkg/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-resty/resty/v2"
	"github.com/lab259/cors"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/valyala/fasthttp"
	"gopkg.in/mgo.v2/bson"
)

// Start 启动监听服务
func Start(addr string) {
	handler := cors.Default().Handler(requestHandler)

	ln1, err := net.Listen("tcp4", "0.0.0.0"+addr)
	if err != nil {
		panic(err)
	}

	ln2, err := net.Listen("tcp6", "[::]"+addr)
	if err != nil {
		panic(err)
	}

	s := &fasthttp.Server{
		Handler:            handler,
		MaxRequestBodySize: 1024 * 1024 * 300,
	}

	go func() {
		err := s.Serve(ln1)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		err := s.Serve(ln2)
		if err != nil {
			panic(err)
		}
	}()

	// if err = fasthttp.ListenAndServe(addr, handler); err != nil {
	// 	glog.Fatalf("Error in ListenAndServe: %s", err)
	// }
}

func getIP(ctx *fasthttp.RequestCtx) (ip string) {
	ip = string(ctx.Request.Header.Peek("X-Forwarded-For"))
	if ip == "" {
		ip = ctx.RemoteIP().String()
	}
	return
}

func fooHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Request method is %q\n", ctx.Method())
	fmt.Fprintf(ctx, "RequestURI is %q\n", ctx.RequestURI())
	fmt.Fprintf(ctx, "Requested path is %q\n", ctx.Path())
	fmt.Fprintf(ctx, "Host is %q\n", ctx.Host())
	fmt.Fprintf(ctx, "Query string is %q\n", ctx.QueryArgs())
	fmt.Fprintf(ctx, "User-Agent is %q\n", ctx.UserAgent())
	fmt.Fprintf(ctx, "Connection has been established at %s\n", ctx.ConnTime())
	fmt.Fprintf(ctx, "Request has been started at %s\n", ctx.Time())
	fmt.Fprintf(ctx, "Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	fmt.Fprintf(ctx, "Your ip is %q\n\n", ctx.RemoteIP())
	fmt.Fprintf(ctx, "header X-Forwarded-For %q\n\n", ctx.Request.Header.Peek("X-Forwarded-For"))
}

func barHandler(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)
}

func logHandler(ctx *fasthttp.RequestCtx) {
	glog.Debugf("Request method is %q\n", ctx.Method())
	glog.Debugf("RequestURI is %q\n", ctx.RequestURI())
	glog.Debugf("Requested path is %q\n", ctx.Path())
	glog.Debugf("Host is %q\n", ctx.Host())
	glog.Debugf("Query string is %q\n", ctx.QueryArgs())
	glog.Debugf("User-Agent is %q\n", ctx.UserAgent())
	glog.Debugf("Connection has been established at %s\n", ctx.ConnTime())
	glog.Debugf("Request has been started at %s\n", ctx.Time())
	glog.Debugf("Serial request number for the current connection is %d\n", ctx.ConnRequestNum())
	glog.Debugf("Your ip is %q\n\n", ctx.RemoteIP())
	glog.Debugf("Raw request is:\n---CUT---\n%s\n---CUT---", &ctx.Request)
}

// 短信验证码
func smsHandler(ctx *fasthttp.RequestCtx) {
	phone := string(ctx.QueryArgs().Peek("phone"))
	glog.Debugf("phone %s", phone)
	// TODO 屏蔽先
	if !utils.PhoneRegexp(phone) {
		// fmt.Fprintf(ctx, "%s", "1")
		// return
	}
	//生成
	arg := new(pb.SmscodeRegist)
	arg.Phone = phone
	arg.Type = 1
	arg.Ipaddr = getIP(ctx)
	nodePid.Tell(arg)
	//fmt.Fprintf(ctx, "%s", "0")
}

func setkvstore(ctx *fasthttp.RequestCtx) {
	k := string(ctx.QueryArgs().Peek("k"))
	v := string(ctx.QueryArgs().Peek("v"))

	lk := len(k)
	lv := len(v)
	if lk > 10240 || lv > 10240 {
		glog.Error("setkvstore too large")
		ctx.Error("setkvstore too large", fasthttp.StatusBadRequest)
		return
	}

	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := rdb.Set(c, fmt.Sprintf("%s%s", RDB_KVSTORE_KEY, k), v, 0).Err()
	if err != nil {
		glog.Error("setkvstore error:", err)
		ctx.Error("setkvstore failed", fasthttp.StatusBadRequest)
		return
	}

	fmt.Fprint(ctx, "")
}

func getkvstore(ctx *fasthttp.RequestCtx) {
	k := string(ctx.QueryArgs().Peek("k"))
	c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	v, err := rdb.Get(c, fmt.Sprintf("%s%s", RDB_KVSTORE_KEY, k)).Result()
	if err != nil {
		glog.Error("getkvstore error:", err)
		ctx.Error("getkvstore failed", fasthttp.StatusBadRequest)
		return
	}

	fmt.Fprintf(ctx, "%s", v)
}

// TODO version rule, json格式返回,包含状态或规则
func gateHandler(ctx *fasthttp.RequestCtx) {
	r := "gate.node"
	v := string(ctx.QueryArgs().Peek("version"))
	if v == "" {
		r += "1"
	} else {
		r += v
	}

	glog.Debugf("gate node : %s", r)
	sec1, err1 := cfg.GetSection(r)
	if err1 != nil {
		glog.Error("Unknwon version ", err1)
		ctx.Error("Unknwon version", fasthttp.StatusBadRequest)
		return
	}
	key2, err2 := sec1.GetKey("host")
	if err2 != nil {
		glog.Error("Unknwon version ", err2)
		ctx.Error("Unknwon version", fasthttp.StatusBadRequest)
		return
	}

	// var scheme string
	// if env == "dev" {
	// 	scheme = "ws"
	// } else {
	// 	scheme = "wss"
	// }

	u := url.URL{Scheme: "ws", Host: key2.Value(), Path: "/"}
	host := u.String()
	host = strings.Replace(host, "%2F", "/", -1)
	logHandler(ctx)
	glog.Debugf("gate host %s", host)
	if aesStatus {
		host = string(aesEn(host))
	}

	if env == "pro" {
		key3, err3 := sec1.GetKey("host2")
		if err3 != nil {
			glog.Error("Unknwon version ", err3)
			ctx.Error("Unknwon version", fasthttp.StatusBadRequest)
			return
		}
		u1 := url.URL{Scheme: "wss", Host: key3.Value(), Path: "/"}
		host1 := u1.String()
		host1 = strings.Replace(host1, "%2F", "/", -1)
		host += ";" + host1
	}

	fmt.Fprintf(ctx, "%s", string(host))
}

func guestHandler(ctx *fasthttp.RequestCtx) {
	defer func() {
		if err := recover(); err != nil {
			glog.Error(err)
		}
	}()
	result := make(map[string]int)
	result["status"] = 0
	switch string(ctx.Method()) {
	case "GET":
	default:
		body, _ := json.Marshal(result)
		fmt.Fprintf(ctx, "%s", string(body))
		return
	}
	deviceid := ctx.QueryArgs().Peek("deviceid")
	count := data.Count(data.PlayerUsers, bson.M{"ad__device_id": string(deviceid), "phone": bson.M{"$ne": ""}})
	result["status"] = count
	body, _ := json.Marshal(result)
	fmt.Fprintf(ctx, "%s", string(body))
}

// 获取客服地址
func customerService(ctx *fasthttp.RequestCtx) {
	result := map[string]string{
		"mail":     "",
		"telegram": "",
		"whats":    "",
	}

	defer func() {
		body, _ := json.Marshal(result)
		fmt.Fprintf(ctx, "%s", string(body))
	}()

	switch string(ctx.Method()) {
	case "GET":
	default:
		return
	}

	addrs := config.GetCustomerAddressMap()
	for _, addr := range addrs {
		result["mail"] = addr.Mail
		result["telegram"] = addr.Telegram
		result["whats"] = addr.WhatsApp
		break
	}
	// body, _ := json.Marshal(result)
	// fmt.Fprintf(ctx, "%s", string(body))
}

func clientlog(ctx *fasthttp.RequestCtx) {
	deviceId := string(ctx.QueryArgs().Peek("deviceId"))
	time := string(ctx.QueryArgs().Peek("time"))
	content := string(ctx.QueryArgs().Peek("content"))

	clog := data.ClientLog{
		DeviceId: deviceId,
		Time:     time,
		Content:  content,
	}
	clog.Save()
	fmt.Fprintf(ctx, "%s", "ok")
}

func grabber(ctx *fasthttp.RequestCtx) {
	fmt.Fprintf(ctx, "%s", "ok")
}

func uploadPhoto(ctx *fasthttp.RequestCtx) {
	AccessKeyID := "AKIA4MVEDDUHKBIWLAFD"
	SecretAccessKey := "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"
	S3Region := "ap-southeast-1"
	S3Bucket := "bucket-indiagame-adjust"

	userid := string(ctx.FormValue("userid"))
	if userid == "" {
		ctx.Error("userid is empty", fasthttp.StatusBadRequest)
		return
	}

	dir := "photoProduct"
	if env == "dev" {
		dir = "photoDev"
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	size := fileHeader.Size

	buffer := make([]byte, size)
	f.Read(buffer)

	creds := credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, "")
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(S3Region),
		Credentials: creds,
	})
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	lastIndex := strings.LastIndex(fileHeader.Filename, ".")
	if lastIndex < 0 {
		ctx.Error("photo format error", fasthttp.StatusBadRequest)
		return
	}

	suffix := fileHeader.Filename[lastIndex:]
	fileName := fmt.Sprintf("%s_%d%s", userid, time.Now().UnixMilli(), suffix)

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(S3Bucket),
		Key:                aws.String(dir + "/" + fileName),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	fmt.Fprintf(ctx, "%s", "ok")

	// 写db
	nodePid.Tell(&pb.UploadPhotoLog{Userid: userid, Url: cdnUrl + dir + "/" + fileName})
}

func uploadimg(ctx *fasthttp.RequestCtx) {
	AccessKeyID := "AKIA4MVEDDUHKBIWLAFD"
	SecretAccessKey := "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"
	S3Region := "ap-southeast-1"
	S3Bucket := "bucket-indiagame-adjust"

	dir := string(ctx.FormValue("dir"))

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	size := fileHeader.Size

	buffer := make([]byte, size)
	f.Read(buffer)

	creds := credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, "")
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(S3Region),
		Credentials: creds,
	})
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(S3Bucket),
		Key:                aws.String(dir + "/" + fileHeader.Filename),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	fmt.Fprintf(ctx, "%s", "ok")
}

func uploadutr(ctx *fasthttp.RequestCtx) {
	AccessKeyID := "AKIA4MVEDDUHKBIWLAFD"
	SecretAccessKey := "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"
	S3Region := "ap-southeast-1"
	S3Bucket := "bucket-indiagame-adjust"

	id := string(ctx.FormValue("id"))
	if id == "" {
		ctx.Error("id is empty", fasthttp.StatusBadRequest)
		return
	}

	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		ctx.Error("id format error", fasthttp.StatusBadRequest)
		return
	}

	uid := parts[0]
	orderid := parts[1]

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	filename, err := gonanoid.New() // 默认生成21位字符
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	filename = fmt.Sprintf("%s_%s_%s", uid, orderid, filename)

	ext := filepath.Ext(fileHeader.Filename)
	uniqueFilename := filename + ext

	f, err := fileHeader.Open()
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	size := fileHeader.Size

	buffer := make([]byte, size)
	f.Read(buffer)

	creds := credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, "")
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(S3Region),
		Credentials: creds,
	})
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(S3Bucket),
		Key:                aws.String("utr" + "/" + uniqueFilename),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	fmt.Fprintf(ctx, "%s", "ok")

	url := "https://bucket-indiagame-adjust.s3.ap-southeast-1.amazonaws.com/utr"

	msg := &pb.UploadUtrReq{
		Uid:      uid,
		OrderId:  orderid,
		FileName: fmt.Sprintf("%s/%s", url, uniqueFilename),
	}

	forwardNode(msg)
}

var (
	PictureExts = []string{".jpg", ".png", ".bmp", ".webp"}
	VideoExts   = []string{".mp4", ".mkv", ".wmv", ".rmvb", ".flv", ".avi"}
	PDFExts     = []string{".pdf"}
)

// 客服聊天文件上传
func uploadCustomer(ctx *fasthttp.RequestCtx) {
	AccessKeyID := "AKIA4MVEDDUHKBIWLAFD"
	SecretAccessKey := "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"
	S3Region := "ap-southeast-1"
	S3Bucket := "bucket-indiagame-adjust"

	userid := string(ctx.FormValue("userid"))
	if userid == "" {
		ctx.Error("userid is empty", fasthttp.StatusBadRequest)
		return
	}

	msg := &pb.CustomerUploadFile{Userid: userid}
	defer forwardNode(msg)

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		msg.Error = fmt.Sprintf("parse upload file error1: %s", err.Error())
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		msg.Error = fmt.Sprintf("parse upload file error2: %s", err.Error())
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	var uniqueFilename, fileName, fileExt string
	fileName = fileHeader.Filename
	fileExt = filepath.Ext(fileHeader.Filename)

	uniqueFilename = fmt.Sprintf("%s_%d%s", userid, time.Now().UnixMilli(), fileExt)

	size := fileHeader.Size

	buffer := make([]byte, size)
	_, err = f.Read(buffer)
	if err != nil {
		msg.Error = fmt.Sprintf("read upload file error3: %s", err.Error())
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	creds := credentials.NewStaticCredentials(AccessKeyID, SecretAccessKey, "")
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(S3Region),
		Credentials: creds,
	})
	if err != nil {
		msg.Error = fmt.Sprintf("connect s3 error4: %s", err.Error())
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}

	_, err = s3.New(s).PutObject(&s3.PutObjectInput{
		Bucket:             aws.String(S3Bucket),
		Key:                aws.String("customer" + "/" + uniqueFilename),
		ACL:                aws.String("private"),
		Body:               bytes.NewReader(buffer),
		ContentLength:      aws.Int64(size),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	})
	if err != nil {
		msg.Error = fmt.Sprintf("upload file to s3 error5: %s", err.Error())
		glog.Error(err)
		ctx.Error(err.Error(), fasthttp.StatusBadRequest)
		return
	}
	fmt.Fprintf(ctx, "%s", "ok")

	url := "https://bucket-indiagame-adjust.s3.ap-southeast-1.amazonaws.com/customer"

	msg.Url = fmt.Sprintf("%s/%s", url, uniqueFilename)
	msg.FileName = fileName
	msg.FileExt = fileExt
	msg.FileSize = size
	if utils.SliceIn(fileExt, PictureExts...) {
		msg.Ctype = 1
	} else if utils.SliceIn(fileExt, VideoExts...) {
		msg.Ctype = 2
	} else if utils.SliceIn(fileExt, PDFExts...) {
		msg.Ctype = 3
	} else {
		msg.Ctype = 4
	}
}

// 大厅轮播图地址
func hallbanner(ctx *fasthttp.RequestCtx) {
	banners := data.GetHallBannerList()
	if banners == nil {
		fmt.Fprintf(ctx, "%s", "[]")
		return
	}
	s, err := json.Marshal(banners)
	if err != nil {
		glog.Error("hall banner err:", err)
		fmt.Fprintf(ctx, "%s", "[]")
		return
	}
	fmt.Fprintf(ctx, "%s", s)
}

// bonus中心banner地址
func bonusbanner(ctx *fasthttp.RequestCtx) {
	banners := data.GetBonusBannerList()
	if banners == nil {
		fmt.Fprintf(ctx, "%s", "[]")
		return
	}
	s, err := json.Marshal(banners)
	if err != nil {
		glog.Error("hall banner err:", err)
		fmt.Fprintf(ctx, "%s", "[]")
		return
	}
	fmt.Fprintf(ctx, "%s", s)
}

// 获取游戏列表(游戏客户端发起)
func getGameList(ctx *fasthttp.RequestCtx) {
	// type ResponseData struct {
	// 	GameId            int    `json:"gameId"`
	// 	GameName          string `json:"gameName"`
	// 	Status            int    `json:"status"`
	// 	GameType          string `json:"gameType"`
	// 	Rtp               string `json:"rtp"`
	// 	Logo              string `json:"logo"`
	// 	CompanyId         string `json:"companyId"`
	// 	HorizontalSupport int    `json:"horizontalSupport"`
	// 	VerticalSupport   int    `json:"verticalSupport"`
	// 	TryGameSupport    int    `json:"tryGameSupport"`
	// }

	// type Response struct {
	// 	Code int             `json:"code"`
	// 	Msg  string          `json:"msg"`
	// 	Data []*ResponseData `json:"data"`
	// }

	// 创建一个 Resty 客户端
	client := resty.New()

	if proxyUrl != "" {
		// 设置代理服务器地址和端口
		client.SetProxy(proxyUrl)
	}

	// 设置请求头，指定 Content-Type 为 application/json
	client.SetHeader("Content-Type", "application/json")

	// 准备要发送的 JSON 数据
	requestBody := map[string]interface{}{
		"accessKey": accessKey,
		"timestamp": time.Now().UnixMilli(),
		"currency":  currency,
	}

	// 发起 POST 请求
	resp, err := client.R().
		SetBody(requestBody).
		Post(fmt.Sprintf("%s/merchant/pub/game/getGameList", apiUrl))

	if err != nil {
		fmt.Fprintf(ctx, "%s", err.Error())
		return
	}

	if resp.StatusCode() != 200 {
		fmt.Fprintf(ctx, "%s", resp.Status())
	}

	fmt.Fprintf(ctx, "%s", resp.String())
}

// 获取游戏地址(游戏客户端发起)
func getGameAddr(ctx *fasthttp.RequestCtx) {
	type ResponseData struct {
		Type int    `json:"type"`
		Data string `json:"data"`
	}

	type Response struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *ResponseData `json:"data"`
	}

	// 玩家类型
	userType := 0
	//参数获取
	gameIdStr := string(ctx.QueryArgs().Peek("gameId"))
	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		fmt.Fprintf(ctx, "%s", err.Error())
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
	if rsp, ok := res.(*pb.GetUserInfoRsp); ok {
		if rsp.Err == "" {
			recharge = rsp.Recharge
			userType = int(rsp.UserType)
		} else {
			fmt.Fprintf(ctx, "%s", rsp.Err)
			return
		}
	} else {
		fmt.Fprintf(ctx, "%s", "call node error")
		return
	}
	_ = userType
	_ = recharge

	// if recharge < 200*100 && userType == 0 {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "You need to recharge at least 200 to play this game."})
	// 	ctx.Write(data)
	// 	// fmt.Fprintf(ctx, "%s", "200")
	// 	return
	// }

	// // B类充200
	// if recharge < 200*100 && userType == 1 {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "You need to recharge at least 200 to play this game."})
	// 	ctx.Write(data)
	// 	// fmt.Fprintf(ctx, "%s", "200")
	// 	return
	// }

	// 创建一个 Resty 客户端
	client := resty.New()

	if proxyUrl != "" {
		// 设置代理服务器地址和端口
		client.SetProxy(proxyUrl)
	}

	// 设置请求头，指定 Content-Type 为 application/json
	client.SetHeader("Content-Type", "application/json")

	// 准备要发送的 JSON 数据
	requestBody := map[string]interface{}{
		"accessKey": accessKey,
		"gameId":    gameId,
		"userToken": userToken,
		"currency":  currency,
		"clientIp":  ctx.RemoteIP().String(),
		"timestamp": time.Now().UnixMilli(),
		"lang":      "en",
	}

	// 发起 POST 请求
	resp, err := client.R().
		SetBody(requestBody).
		Post(fmt.Sprintf("%s/merchant/pub/game/getGameAddr", apiUrl))

	if err != nil {
		fmt.Fprintf(ctx, "%s", err.Error())
		return
	}

	if resp.StatusCode() != 200 {
		fmt.Fprintf(ctx, "%s", resp.Status())
	}

	fmt.Fprintf(ctx, "%s", resp.String())
}

// 验证用户token（外接调用）
func checkToken(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")

	type ResponseData struct {
		UserId    string `json:"userId"`
		Currency  string `json:"currency"`
		Balance   string `json:"balance"`
		NickName  string `json:"nickName"`
		AvatarUrl string `json:"avatarUrl"`
	}

	type Response struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *ResponseData `json:"data"`
	}

	type Request struct {
		AccessKey string `json:"accessKey"`
		UserToken string `json:"userToken"`
		Currency  string `json:"currency"`
		Timestamp int64  `json:"timestamp"`
	}

	//参数获取
	body := ctx.PostBody()

	var req Request
	err := json.Unmarshal(body, &req)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}

	//校验参数
	if req.AccessKey == "" || req.UserToken == "" || req.Currency == "" || req.Timestamp < 0 {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "param error"})
		ctx.Write(data)
		return
	}

	if req.AccessKey != accessKey {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "accessKey error"})
		ctx.Write(data)
		return
	}

	if req.Currency != currency {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "currency error"})
		ctx.Write(data)
		return
	}

	//token验证
	uid, err := handler.Valid(req.UserToken)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "token invalid"})
		ctx.Write(data)
		return
	}

	msg := &pb.GetUserInfoReq{Uid: uid}
	res, err := callNode(msg)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "request error"})
		ctx.Write(data)
		return
	}

	if rsp, ok := res.(*pb.GetUserInfoRsp); ok {
		if rsp.Err == "" {
			balance := float64(rsp.Balance) / 100.0
			balanceStr := strconv.FormatFloat(balance, 'f', 6, 64)

			data, _ := json.Marshal(&Response{
				Code: 200,
				Msg:  "success",
				Data: &ResponseData{
					UserId:    uid,
					Currency:  currency,
					Balance:   balanceStr,
					NickName:  rsp.NickName,
					AvatarUrl: fmt.Sprintf("%s/img_%s.png", headUrl, rsp.Photo),
				}})
			ctx.Write(data)
			return
		} else {
			data, _ := json.Marshal(&Response{Code: 400, Msg: rsp.Err})
			ctx.Write(data)
			return
		}
	} else {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "data error"})
		ctx.Write(data)
		return
	}

}

// 查询用户余额（外接调用）
func getBalance(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")

	type ResponseData struct {
		UserId   string `json:"userId"`
		Currency string `json:"currency"`
		Balance  string `json:"balance"`
	}

	type Response struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *ResponseData `json:"data"`
	}

	type Request struct {
		AccessKey string `json:"accessKey"`
		UserToken string `json:"userToken"`
		Currency  string `json:"currency"`
		Timestamp int64  `json:"timestamp"`
	}

	//参数获取
	body := ctx.PostBody()

	var req Request
	err := json.Unmarshal(body, &req)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}

	//校验参数
	if req.AccessKey == "" || req.UserToken == "" || req.Currency == "" || req.Timestamp < 0 {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "param error"})
		ctx.Write(data)
		return
	}

	if req.AccessKey != accessKey {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "accessKey error"})
		ctx.Write(data)
		return
	}

	if req.Currency != currency {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "currency error"})
		ctx.Write(data)
		return
	}

	//token验证
	uid, err := handler.Valid(req.UserToken)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "token invalid"})
		ctx.Write(data)
		return
	}

	msg := &pb.GetUserInfoReq{Uid: uid}
	res, err := callNode(msg)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "request error"})
		ctx.Write(data)
		return
	}

	if rsp, ok := res.(*pb.GetUserInfoRsp); ok {
		if rsp.Err == "" {
			balance := float64(rsp.Balance) / 100.0
			balanceStr := strconv.FormatFloat(balance, 'f', 6, 64)

			data, _ := json.Marshal(&Response{
				Code: 200,
				Msg:  "success",
				Data: &ResponseData{
					UserId:   uid,
					Currency: currency,
					Balance:  balanceStr,
				}})
			ctx.Write(data)
			return
		} else {
			data, _ := json.Marshal(&Response{Code: 400, Msg: rsp.Err})
			ctx.Write(data)
			return
		}
	} else {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "data error"})
		ctx.Write(data)
		return
	}
}

// 游戏投注
func bet(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")

	type ResponseData struct {
		MerchantOrderNo string `json:"merchantOrderNo"`
		OrderNo         string `json:"orderNo"`
		Currency        string `json:"currency"`
		Balance         string `json:"balance"`
	}

	type Response struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *ResponseData `json:"data"`
	}

	type Request struct {
		AccessKey string `json:"accessKey"`
		UserId    string `json:"userId"`
		GameId    int32  `json:"gameId"`
		RoundId   string `json:"roundId"`
		OrderNo   string `json:"orderNo"`
		Amount    string `json:"amount"`
		Currency  string `json:"currency"`
		Timestamp int64  `json:"timestamp"`
	}

	//参数获取
	body := ctx.PostBody()

	var req Request
	err := json.Unmarshal(body, &req)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}

	//校验参数
	if req.AccessKey == "" || req.UserId == "" || req.GameId == 0 || req.RoundId == "" || req.OrderNo == "" || req.Amount == "" || req.Currency == "" || req.Timestamp < 0 {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "param error"})
		ctx.Write(data)
		return
	}

	if req.AccessKey != accessKey {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "accessKey error"})
		ctx.Write(data)
		return
	}

	if req.Currency != currency {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "currency error"})
		ctx.Write(data)
		return
	}

	//token验证
	// uid, err := handler.Valid(req.UserToken)
	// if err != nil {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "token invalid"})
	// 	ctx.Write(data)
	// 	return
	// }

	//下注额处理
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}
	amount = amount * 100
	amountInt64 := int64(amount)
	// if amountInt64 == 0 {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "bet amount precision error"})
	// 	ctx.Write(data)
	// 	return
	// }

	msg := &pb.ExternalBetReq{
		Uid:     req.UserId,
		GameId:  req.GameId,
		RoundId: req.RoundId,
		OrderNo: req.OrderNo,
		Amount:  amountInt64,
	}
	res, err := callNode(msg)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "request error"})
		ctx.Write(data)
		return
	}

	if rsp, ok := res.(*pb.ExternalBetRsp); ok {
		if rsp.Err == "" {
			balance := float64(rsp.Balance) / 100.0
			balanceStr := strconv.FormatFloat(balance, 'f', 6, 64)

			data, _ := json.Marshal(&Response{
				Code: 200,
				Msg:  "success",
				Data: &ResponseData{
					MerchantOrderNo: rsp.MerchantOrderNo,
					OrderNo:         rsp.OrderNo,
					Currency:        currency,
					Balance:         balanceStr,
				}})
			ctx.Write(data)
			err = mq.NatsPublish(mq.TopicOnlineUser, &pb.OnlineUserTTL{
				Userid: req.UserId,
				GameId: strconv.Itoa(int(req.GameId)),
			})
			if err != nil {
				glog.Error(err)
			}
			return
		} else {
			data, _ := json.Marshal(&Response{Code: int(rsp.ErrCode), Msg: rsp.Err})
			ctx.Write(data)
			return
		}
	} else {
		data, _ := json.Marshal(&Response{Code: 500, Msg: "data error"})
		ctx.Write(data)
		return
	}
}

// 游戏结算
func reward(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")

	type ResponseData struct {
		MerchantOrderNo string `json:"merchantOrderNo"`
		OrderNo         string `json:"orderNo"`
		Currency        string `json:"currency"`
		Balance         string `json:"balance"`
	}

	type Response struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *ResponseData `json:"data"`
	}

	type Request struct {
		AccessKey    string `json:"accessKey"`
		UserId       string `json:"userId"`
		GameId       int32  `json:"gameId"`
		RoundId      string `json:"roundId"`
		OrderNo      string `json:"orderNo"`
		RewardAmount string `json:"rewardAmount"`
		Currency     string `json:"currency"`
		Timestamp    int64  `json:"timestamp"`
	}

	//参数获取
	body := ctx.PostBody()

	var req Request
	err := json.Unmarshal(body, &req)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}

	//校验参数
	if req.AccessKey == "" || req.UserId == "" || req.GameId == 0 || req.RoundId == "" || req.OrderNo == "" || req.RewardAmount == "" || req.Currency == "" || req.Timestamp < 0 {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "param error"})
		ctx.Write(data)
		return
	}

	if req.AccessKey != accessKey {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "accessKey error"})
		ctx.Write(data)
		return
	}

	if req.Currency != currency {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "currency error"})
		ctx.Write(data)
		return
	}

	//token验证
	// uid, err := handler.Valid(req.UserToken)
	// if err != nil {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "token invalid"})
	// 	ctx.Write(data)
	// 	return
	// }

	//下注额处理
	rewardAmount, err := strconv.ParseFloat(req.RewardAmount, 64)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}
	rewardAmount = rewardAmount * 100
	rewardAmountInt64 := int64(rewardAmount)
	// if rewardAmountInt64 == 0 {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "bet amount precision error"})
	// 	ctx.Write(data)
	// 	return
	// }

	msg := &pb.ExternalRewardReq{
		Uid:          req.UserId,
		GameId:       req.GameId,
		RoundId:      req.RoundId,
		OrderNo:      req.OrderNo,
		RewardAmount: rewardAmountInt64,
	}
	res, err := callNode(msg)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "request error"})
		ctx.Write(data)
		return
	}

	if rsp, ok := res.(*pb.ExternalRewardRsp); ok {
		if rsp.Err == "" {
			balance := float64(rsp.Balance) / 100.0
			balanceStr := strconv.FormatFloat(balance, 'f', 6, 64)

			data, _ := json.Marshal(&Response{
				Code: 200,
				Msg:  "success",
				Data: &ResponseData{
					MerchantOrderNo: rsp.MerchantOrderNo,
					OrderNo:         rsp.OrderNo,
					Currency:        currency,
					Balance:         balanceStr,
				}})
			ctx.Write(data)
			return
		} else {
			data, _ := json.Marshal(&Response{Code: 400, Msg: rsp.Err})
			ctx.Write(data)
			return
		}
	} else {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "data error"})
		ctx.Write(data)
		return
	}
}

// 取消订单
func cancel(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")

	type ResponseData struct {
		MerchantOrderNo string `json:"merchantOrderNo"`
		OrderNo         string `json:"orderNo"`
		Currency        string `json:"currency"`
		Balance         string `json:"balance"`
	}

	type Response struct {
		Code int           `json:"code"`
		Msg  string        `json:"msg"`
		Data *ResponseData `json:"data"`
	}

	type Request struct {
		Type              int32  `json:"type"`
		AccessKey         string `json:"accessKey"`
		UserId            string `json:"userId"`
		GameId            int32  `json:"gameId"`
		RoundId           string `json:"roundId"`
		OrderNo           string `json:"orderNo"`
		CancelOrderNo     string `json:"cancelOrderNo"`
		CancelPlatOrderNo string `json:"cancelPlatOrderNo"`
		OrderAmount       string `json:"orderAmount"`
		Currency          string `json:"currency"`
		OrderDesc         string `json:"orderDesc"`
		Timestamp         int64  `json:"timestamp"`
	}

	//参数获取
	body := ctx.PostBody()

	var req Request
	err := json.Unmarshal(body, &req)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: err.Error()})
		ctx.Write(data)
		return
	}

	//校验参数
	if req.Type == 0 || req.AccessKey == "" || req.UserId == "" || req.GameId == 0 || req.RoundId == "" || req.OrderNo == "" ||
		req.CancelOrderNo == "" || req.CancelPlatOrderNo == "" || req.OrderAmount == "" || req.Currency == "" || req.OrderDesc == "" || req.Timestamp < 0 {
		data, _ := json.Marshal(&Response{Code: 500, Msg: "param error"})
		ctx.Write(data)
		return
	}

	if req.AccessKey != accessKey {
		data, _ := json.Marshal(&Response{Code: 500, Msg: "accessKey error"})
		ctx.Write(data)
		return
	}

	if req.Currency != currency {
		data, _ := json.Marshal(&Response{Code: 500, Msg: "currency error"})
		ctx.Write(data)
		return
	}

	//token验证
	// uid, err := handler.Valid(req.UserToken)
	// if err != nil {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "token invalid"})
	// 	ctx.Write(data)
	// 	return
	// }

	//下注额处理
	orderAmount, err := strconv.ParseFloat(req.OrderAmount, 64)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 500, Msg: err.Error()})
		ctx.Write(data)
		return
	}
	orderAmount = orderAmount * 100
	orderAmountInt64 := int64(orderAmount)
	// if rewardAmountInt64 == 0 {
	// 	data, _ := json.Marshal(&Response{Code: 400, Msg: "bet amount precision error"})
	// 	ctx.Write(data)
	// 	return
	// }

	msg := &pb.ExternalCancelReq{
		Type:              req.Type,
		Uid:               req.UserId,
		GameId:            req.GameId,
		RoundId:           req.RoundId,
		OrderNo:           req.OrderNo,
		CancelOrderNo:     req.CancelOrderNo,
		CancelPlatOrderNo: req.CancelPlatOrderNo,
		OrderAmount:       orderAmountInt64,
	}
	res, err := callNode(msg)
	if err != nil {
		data, _ := json.Marshal(&Response{Code: 400, Msg: "request error"})
		ctx.Write(data)
		return
	}

	if rsp, ok := res.(*pb.ExternalCancelRsp); ok {
		if rsp.Err == "" {
			balance := float64(rsp.Balance) / 100.0
			balanceStr := strconv.FormatFloat(balance, 'f', 6, 64)

			data, _ := json.Marshal(&Response{
				Code: 200,
				Msg:  "success",
				Data: &ResponseData{
					MerchantOrderNo: rsp.MerchantOrderNo,
					OrderNo:         rsp.OrderNo,
					Currency:        currency,
					Balance:         balanceStr,
				}})
			ctx.Write(data)
			return
		} else {
			data, _ := json.Marshal(&Response{Code: 500, Msg: rsp.Err})
			ctx.Write(data)
			return
		}
	} else {
		data, _ := json.Marshal(&Response{Code: 500, Msg: "data error"})
		ctx.Write(data)
		return
	}
}

func adjust(ctx *fasthttp.RequestCtx) {
	app_token := string(ctx.QueryArgs().Peek("app_token"))
	app_version := string(ctx.QueryArgs().Peek("app_version"))
	app_id := string(ctx.QueryArgs().Peek("app_id"))
	environment := string(ctx.QueryArgs().Peek("environment"))
	tracker_name := string(ctx.QueryArgs().Peek("tracker_name"))
	last_tracker_name := string(ctx.QueryArgs().Peek("last_tracker_name"))
	network_name := string(ctx.QueryArgs().Peek("network_name"))
	campaign_name := string(ctx.QueryArgs().Peek("campaign_name"))
	adgroup_name := string(ctx.QueryArgs().Peek("adgroup_name"))
	creative_name := string(ctx.QueryArgs().Peek("creative_name"))
	installed_at := string(ctx.QueryArgs().Peek("installed_at"))
	reinstalled_at := string(ctx.QueryArgs().Peek("reinstalled_at"))
	uninstalled_at := string(ctx.QueryArgs().Peek("uninstalled_at"))
	adid := string(ctx.QueryArgs().Peek("adid"))
	gps_adid := string(ctx.QueryArgs().Peek("gps_adid"))
	android_id := string(ctx.QueryArgs().Peek("android_id"))
	sdk_version := string(ctx.QueryArgs().Peek("sdk_version"))
	language := string(ctx.QueryArgs().Peek("language"))
	city := string(ctx.QueryArgs().Peek("city"))
	device_type := string(ctx.QueryArgs().Peek("device_type"))
	device_name := string(ctx.QueryArgs().Peek("device_name"))
	ip_address := string(ctx.QueryArgs().Peek("ip_address"))
	proxy_ip_address := string(ctx.QueryArgs().Peek("proxy_ip_address"))
	user_agent := string(ctx.QueryArgs().Peek("user_agent"))

	ad := data.Adjust{
		AppToken:        app_token,
		AppVersion:      app_version,
		AppId:           app_id,
		Environment:     environment,
		TrackerName:     tracker_name,
		LastTrackerName: last_tracker_name,
		NetworkName:     network_name,
		CampaignName:    campaign_name,
		AdgroupName:     adgroup_name,
		CreativeName:    creative_name,
		InstalledAt:     installed_at,
		ReinstalledAt:   reinstalled_at,
		UninstalledAt:   uninstalled_at,
		Adid:            adid,
		GpsAdid:         gps_adid,
		AndroidId:       android_id,
		SdkVersion:      sdk_version,
		Language:        language,
		City:            city,
		DeviceType:      device_type,
		DeviceName:      device_name,
		IpAddress:       ip_address,
		ProxyIpAddress:  proxy_ip_address,
		UserAgent:       user_agent,
	}
	ad.Save()
	fmt.Fprintf(ctx, "%s", "ok")
}

// 接受网页端adid
func receivePcAdjust(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Method()) {
	case "POST":
	default:
		fmt.Fprintf(ctx, "%s", "fail")
		return
	}
	userid := string(ctx.QueryArgs().Peek("userid"))
	adid := string(ctx.QueryArgs().Peek("adid"))
	if userid == "" || adid == "" {
		glog.Errorf("web page err adid:%s, userid:%s", adid, userid)
		fmt.Fprintf(ctx, "%s", "fail")
		return
	}
	msg := &pb.WebPageAdid{
		Adid:   adid,
		Userid: userid,
	}
	rsp, err := callNode(msg)
	if err != nil {
		fmt.Fprintf(ctx, "%s", "fail")
		return
	}
	if res, ok := rsp.(*pb.WebPageAdided); ok {
		if res.Code == 200 {
			fmt.Fprintf(ctx, "%s", "success")
			return
		} else {
			fmt.Fprintf(ctx, "%s", "fail")
			return
		}

	}
	fmt.Fprintf(ctx, "%s", "success")
}

func playsharetest(ctx *fasthttp.RequestCtx) {
	// if env != "dev" {
	// 	fmt.Fprintf(ctx, "%s", "fail")
	// 	return
	// }

	switch string(ctx.Method()) {
	case "GET":
	default:
		fmt.Fprintf(ctx, "%s", "fail")
		return
	}
	deviceid := ctx.QueryArgs().Peek("deviceid")
	userid := ctx.QueryArgs().Peek("userid")
	msg := &pb.PlayshareTest{Userid: string(userid), Deviceid: string(deviceid)}
	callNode(msg)
}

func requestHandler(ctx *fasthttp.RequestCtx) {
	defer ctx.SetConnectionClose()
	switch string(ctx.Path()) {
	case "/api/web":
		webHandler(ctx)
	case "/api/webjson":
		webJSONHandler(ctx)
	case "/api/sms":
		smsHandler(ctx)
	case "/api/guest":
		guestHandler(ctx)
	case "/api/gate":
		gateHandler(ctx)
	case "/api/setkvstore":
		setkvstore(ctx)
	case "/api/getkvstore":
		getkvstore(ctx)
	// case "/api/wxpay/notice":
	// 	wxpayHandler(ctx)
	// case "/api/jtpay/order":
	// 	jtpayOrder(ctx)
	// case "/api/jtpay/return":
	// 	jtpayReturn(ctx)
	// case "/api/jtpay/notify":
	// 	jtpayNotify(ctx)
	case "/api/login/customer":
		customerService(ctx)
	// case "/api/wxmp/oauth2":
	// 	wxmpOauth2(ctx)
	case "/api/wxmp/shorturl":
		// wxmpShortURL(ctx)
	case "/api/wxmp/qrcode":
		wxmpQRcode(ctx)
	case "/api/download":
		//download(ctx)
		downloadPageHandler(ctx)
	//case "/wxmp/wx":
	case "/api/foo":
		fooHandler(ctx)
	case "/api/bar":
		barHandler(ctx)
	case "/api/adjust":
		adjust(ctx)
	case "/api/clientlog":
		clientlog(ctx)
	case "/api/grabber":
		grabber(ctx)
	case "/api/uploadimg":
		uploadimg(ctx)
	case "/api/uploadphoto":
		uploadPhoto(ctx)
	case "/api/uploadutr":
		uploadutr(ctx)
	case "/api/uploadcustomer":
		uploadCustomer(ctx)
	case "/api/hallbanner":
		hallbanner(ctx)
	case "/api/bonusbanner":
		bonusbanner(ctx)
	case "/api/getGameList":
		getGameList(ctx)
	case "/api/getGameAddr":
		getGameAddr(ctx)
	case "/game/user/checkToken":
		checkToken(ctx)
	case "/game/user/getBalance":
		getBalance(ctx)
	case "/game/user/bet":
		bet(ctx)
	case "/game/user/reward":
		reward(ctx)
	case "/api/pcadjust":
		receivePcAdjust(ctx)
	case "/api/playshare":
		playsharetest(ctx)
	case "/game/user/cancel":
		cancel(ctx)
	case "/ping/dbms":
		heartBeat(ctx)
	default:
		payRequestHandler(ctx)
	}
}

// 代收回调
func payRequestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/pay/notify":
		payNotify(ctx, 1)
	case "/api/withdraw/notify":
		payNotify(ctx, 2)
	case "/api/withdraw/repeatWithdraw":
		repeatWithdraw(ctx)
	default:
		smsRequestHandler(ctx)
	}
}

// 短信回调
func smsRequestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	// case "/api/smsnotify/kmi":
	// 	kmiSms(ctx)
	// case "/api/smsnotify/mt":
	// 	mtSms(ctx)
	// case "/api/smsnotify/xxy":
	// 	xxySms(ctx)
	default:
		efiRequestHandler(ctx)
	}
}

// 视讯回调
func efiRequestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/api/efiGameList":
		efiGameList(ctx)
	case "/api/efiLaunchGame":
		efiLaunchGame(ctx)
	case "/efi/Seamless/GetBalance":
		efiGetBalance(ctx)
	case "/efi/Seamless/PlaceBet":
		efiPlaceBet(ctx)
	case "/efi/Seamless/GameResult":
		efiGameResult(ctx)
	case "/efi/Seamless/Rollback":
		efiRollback(ctx)
	case "/efi/Seamless/efiCancelBet":
		efiCancelBet(ctx)
	case "/efi/Seamless/Bonus":
		efiBonus(ctx)
	case "/efi/Seamless/Jackpot":
		efiJackpot(ctx)
	case "/efi/Seamless/MobileLogin":
		efiMobileLogin(ctx)
	case "/efi/Seamless/BuyIn":
		efiBuyIn(ctx)
	case "/efi/Seamless/BuyOut":
		efiBuyOut(ctx)
	case "/efi/Seamless/PushBet":
		efiPushBet(ctx)
	default:
		glog.Errorf("send error api:%s, status:%d", string(ctx.Path()), fasthttp.StatusNotFound)
	}
}
