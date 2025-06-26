package service

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"time"

	"goserver/gen/pb"
	"goserver/internal/web/stats/app/entity"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
)

var gmHost string
var gmPort string
var gmPath string
var gmKey string
var gmUrl string
var gmCdn string

var adUrl string

var payUrl string

func init() {
	gmHost = beego.AppConfig.String("gm.host")
	gmPort = beego.AppConfig.String("gm.port")
	gmPath = beego.AppConfig.String("gm.path")
	gmKey = beego.AppConfig.String("gm.key")
	gmUrl = beego.AppConfig.String("gm.url")
	gmCdn = beego.AppConfig.String("gm.cdn")

	adUrl = beego.AppConfig.String("ad.url")

	payUrl = beego.AppConfig.String("pay.url")
}

var cstDialer = websocket.Dialer{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Gm操作,角色ID,操作类型,操作物品,操作数量
func Gm(msgName, msg string) (string, error) {
	addr := gmHost + ":" + gmPort
	//fmt.Println("addr -> ", addr)
	u := url.URL{Scheme: "wss", Host: addr, Path: gmPath}
	TimeStr := GmTime()
	Token := GmToke(TimeStr)
	//c, _, err := websocket.DefaultDialer.Dial(u.String(),
	//	http.Header{"Token": {Token}})
	d := cstDialer
	d.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c, _, err := d.Dial(u.String(),
		http.Header{"Token": {Token}})
	if err != nil {
		return "", errors.New(fmt.Sprintf("dial err -> %v", err))
	}
	sign := GmSign(msg, TimeStr)
	msg2 := GmMsg(msg, msgName, sign, TimeStr)
	if c != nil {
		c.WriteMessage(websocket.TextMessage, []byte(msg2))
		defer c.Close()
		_, message, err := c.ReadMessage()
		if err != nil {
			return "", errors.New(fmt.Sprintf("read err -> %v", err))
		}
		resp := new(entity.RespErr)
		err = json.Unmarshal(message, resp)
		if err != nil {
			return "", errors.New(fmt.Sprintf("Unmarshal err -> %v", err))
		}
		if resp.ErrCode != 0 {
			return "", errors.New(fmt.Sprintf("resp.ErrMsgi %s", resp.ErrMsg))
		}
		return resp.Result, nil
	}
	return "", errors.New(fmt.Sprintf("c empty err -> %v", err))
}

// 字符串时间
func GmTime() string {
	Time := utils.Timestamp()
	TimeStr := utils.String(Time)
	return TimeStr
}

// Sign := utils.Md5(Key+Now)
// Token := Sign+Now+RandNum
func GmToke(TimeStr string) string {
	Sign := utils.Md5(gmKey + TimeStr)
	Token := Sign + TimeStr + utils.RandStr(6)
	return Token
}

// Sign := TimeStr + Key + Md5(msg)
func GmSign(msg, TimeStr string) string {
	return utils.Md5(TimeStr + gmKey + utils.Md5(msg))
}

// Timestr|Sign|msg_name|msg
func GmMsg(msg, msgName, sign, TimeStr string) string {
	return TimeStr + "|" + sign + "|" + msgName + "|" + msg
}

// http request

func GmRequest(code pb.WebCode, atype pb.ConfigAtype,
	b interface{}) (interface{}, error) {
	//pack
	body, err1 := gmPack(code, atype, b)
	if err1 != nil {
		return nil, err1
	}
	//request
	result, err3 := doHttpPost(gmUrl, body)
	if err3 != nil {
		return nil, err3
	}
	//unpack
	return gmUnpack(code, result)
}

func gmPack(code pb.WebCode, atype pb.ConfigAtype,
	b interface{}) ([]byte, error) {
	msg := new(pb.WebRequest)
	msg.Code = code
	msg.Atype = atype
	switch b.(type) {
	case []byte:
		msg.Data = b.([]byte)
	case *pb.ChangeCurrency:
		msg2 := b.(*pb.ChangeCurrency)
		result, err2 := msg2.Marshal()
		if err2 != nil {
			return []byte{}, err2
		}
		msg.Data = result
	case *pb.PayCurrency:
		msg2 := b.(*pb.PayCurrency)
		result, err2 := msg2.Marshal()
		if err2 != nil {
			return []byte{}, err2
		}
		msg.Data = result
	case *pb.UploadConfig:
		msg2 := b.(*pb.UploadConfig)
		result, err2 := msg2.Marshal()
		if err2 != nil {
			return []byte{}, err2
		}
		msg.Data = result
	case *pb.ModifyStock:
		msg2 := b.(*pb.ModifyStock)
		result, err2 := msg2.Marshal()
		if err2 != nil {
			return []byte{}, err2
		}
		msg.Data = result
	default:
		result, err2 := json.Marshal(b)
		if err2 != nil {
			return []byte{}, err2
		}
		msg.Data = result
	}
	body, err1 := msg.Marshal()
	if err1 != nil {
		return []byte{}, err1
	}
	return body, nil
}

func gmUnpack(code pb.WebCode, body []byte) (interface{}, error) {
	resp := new(pb.WebResponse)
	err1 := resp.Unmarshal(body)
	if err1 != nil {
		fmt.Println("gmUnpack code ", code, err1)
		return nil, err1
	}
	if resp.ErrCode != 0 || resp.ErrMsg != "" {
		fmt.Println("gmUnpack code ", resp.ErrCode, resp.ErrMsg)
		return nil, fmt.Errorf("ErrCode %d, ErrMsg %s", resp.ErrCode, resp.ErrMsg)
	}
	switch code {
	case pb.WebOnline:
		b := make(map[string]int)
		err2 := json.Unmarshal(resp.Result, &b)
		if err2 != nil {
			fmt.Println("gmUnpack code ", code, err2)
			return nil, err2
		}
		return b, nil
	case pb.WebNumber:
		b := make(map[int]int)
		err2 := json.Unmarshal(resp.Result, &b)
		if err2 != nil {
			fmt.Println("gmUnpack code ", code, err2)
			return nil, err2
		}
		return b, nil
	case pb.WebShop:
		return nil, nil
	case pb.WebEnv:
		return nil, nil
	case pb.WebNotice:
		return nil, nil
	case pb.WebGame:
		return nil, nil
	case pb.WebPayChannel:
		return nil, nil
	case pb.WebSetWithDraw:
		return nil, nil
	case pb.WebWithdraw:
		return nil, nil
	case pb.WebSwitch:
		return nil, nil
	case pb.WebVip:
	case pb.WebBuild:
		return nil, nil
	case pb.WebFeedBack:
		return nil, nil
	case pb.WebBlack:
		return nil, nil
	case pb.WebIpWhite:
		return nil, nil
	case pb.WebGive:
		return nil, nil
	case pb.WebModifyNum:
		return nil, nil
	case pb.WebModifyUser:
		return nil, nil
	case pb.WebModifyStock:
		return nil, nil
	case pb.WebRechareBlack:
		return nil, nil
	case pb.WebOnlineUser:
		b := make([]entity.OnlineUser, 0)
		err2 := json.Unmarshal(resp.Result, &b)
		if err2 != nil {
			fmt.Println("gmUnpack code ", code, err2)
			return nil, err2
		}
		return b, nil
	case pb.WebUploadConfig:
		return nil, nil
	case pb.WebPointControl:
		return nil, nil
	case pb.WebSuperior:
		return nil, nil
	case pb.WebChannel:
		return nil, nil
	case pb.WebServerWhite:
		return nil, nil
	case pb.WebGiveWithdraw:
		return nil, nil
	case pb.WebModifyCustomer:
		return nil, nil
	case pb.WebModifyShare:
		return nil, nil
	case pb.WebCloseServer:
		return nil, nil
	case pb.WebPayCallback:
		return nil, nil
	}
	return nil, fmt.Errorf("unknown code %d", code)
}

// http post
func doHttpPost(targetUrl string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewBuffer(body))
	if err != nil {
		return []byte(""), err
	}
	req.Header.Add("Content-type", "text/plain;charset=UTF-8")

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		TLSHandshakeTimeout: 30 * time.Second,
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return respData, nil
}

func UploadRequest(ImgName string, file multipart.File) error {
	// 创建一个缓冲区来存储 form-data 数据
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 将文件内容复制到文件字段中
	// 创建一个文件字段
	fileWriter, err := writer.CreateFormFile("file", ImgName)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}
	// 完成 form-data 数据的写入
	err = writer.Close()
	if err != nil {
		return err
	}

	uploadurl := "http://" + gmHost + ":" + gmPort + "/api/uploadimg"

	// 创建 POST 请求
	request, err := http.NewRequest("POST", uploadurl, body)
	if err != nil {
		return err

	}

	// 设置请求头
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		TLSHandshakeTimeout: 30 * time.Second,
	}
	client := &http.Client{Transport: transport}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	respData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	resStr := string(respData)
	if resStr == "ok" {
		return nil
	}
	return errors.New(resStr)
}

// Timestr|Sign|msg_name|msg
func UploadUrl(ImgName string) string {
	return gmCdn + "/" + ImgName
}

// http get
func AdGetRequest(b []string) (*entity.AndroidScore, error) {
	// 获取评分
	params := url.Values{}
	strUrl := adUrl + "/v1/getScoreByAdIds"
	if len(b) > 0 {
		for _, item := range b {
			params.Add("ad_id", item)
		}
	}
	result, err3 := doHttpGet(strUrl, params)
	if err3 != nil {
		return nil, err3
	}
	//unpack
	return adUnpack(result)
}

func AdGetCheckstand(b []string) (*entity.AndroidCheckstand, error) {
	params := url.Values{}
	// 收银台统计
	strUrl := adUrl + "/v1/stat/day"
	if len(b) > 0 {
		params.Add("startDay", b[0])
		params.Add("endDay", b[1])
	}
	result, err3 := doHttpGet(strUrl, params)
	if err3 != nil {
		return nil, err3
	}
	return adUnpack1(result)
}

// func adPack(b interface{}) (map[string][]string, error) {

// 	// result, err2 := json.Marshal(b, &values)
// 	// if err2 != nil {
// 	// 	return values, err2
// 	// }
// 	return values, nil
// }

func adUnpack(body []byte) (*entity.AndroidScore, error) {
	resp := new(entity.AndroidScore)
	err := json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("adUnpack code ", err)
		return nil, err
	}
	if resp.Code != 1 || resp.Message != "" {
		fmt.Println("adUnpack code ", resp.Code, resp.Message)
		return nil, fmt.Errorf("ErrCode %d, ErrMsg %s", resp.Code, resp.Message)
	}
	return resp, nil
}

func adUnpack1(body []byte) (*entity.AndroidCheckstand, error) {
	resp := new(entity.AndroidCheckstand)
	err := json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("adUnpack code ", err)
		return nil, err
	}
	if resp.Code != 1 || resp.Message != "" {
		fmt.Println("adUnpack code ", resp.Code, resp.Message)
		return nil, fmt.Errorf("ErrCode %d, ErrMsg %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// http get
func doHttpGet(targetUrl string, params url.Values) ([]byte, error) {
	url := targetUrl + "?" + params.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []byte(""), err
	}
	req.Header.Add("Content-type", "text/plain;charset=UTF-8")

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte(""), err
	}

	return respData, nil
}

// 支付节点get请求
func PayGetRequest(b map[string]string, method string) (*entity.WebResponse, error) {
	// 获取评分
	params := url.Values{}
	targetUrl := payUrl + method
	if len(b) > 0 {
		for i, item := range b {
			params.Add(i, item)
		}
	}
	result, err3 := doHttpGet(targetUrl, params)
	if err3 != nil {
		return nil, err3
	}
	//unpack
	return payUnpack(result)
}

func payUnpack(body []byte) (*entity.WebResponse, error) {
	resp := new(entity.WebResponse)
	err := json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("payUnpack code ", err)
		return nil, err
	}
	if resp.Code != 200 {
		fmt.Println("payUnpack code ", resp.Code, resp.ErrMsg)
		return nil, fmt.Errorf("ErrCode %d, ErrMsg %s", resp.Code, resp.ErrMsg)
	}
	return resp, nil
}
