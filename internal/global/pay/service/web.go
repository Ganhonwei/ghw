package service

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/ini.v1"
)

// var gmHost string
// var gmPort string
// var gmPath string
// var gmKey string
var gmUrl string
var Proxy string

func InitWeb(cfg *ini.File) {
	// gmHost = cfg.Section("").Key("gm.host").String()
	// gmPort = cfg.Section("").Key("gm.port").String()
	// gmPath = cfg.Section("").Key("gm.path").String()
	// gmKey = cfg.Section("").Key("gm.key").String()
	gmUrl = cfg.Section("").Key("gm.url").String()
	Proxy = cfg.Section("").Key("proxy").String()
}

// 通知服务器
func PayNotify(order *data.PayNotify, rtype int) ([]byte, error) {
	//打包post数据
	// dictionary := make(map[string]string)
	// dictionary["mer_orderid"] = fmt.Sprint(order.MerOrderID)
	// dictionary["amount"] = strconv.Itoa(int(order.Amount))
	// dictionary["status"] = strconv.Itoa(int(order.Status))
	// dictionary["orderId"] = fmt.Sprint(order.OrderID)
	// dictionary["timestamp"] = fmt.Sprint(order.Timestamp)

	// body := dictionary
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(order)
	//fmt.Printf("body %s, order %#v\n", body, order)
	reqUrl := ""
	if rtype == 1 {
		// 支付
		reqUrl = fmt.Sprintf("%s%s", gmUrl, "/pay/notify")
	} else if rtype == 2 {
		// 提现
		reqUrl = fmt.Sprintf("%s%s", gmUrl, "/withdraw/notify")

		// if order.Status == data.WithdrawSuccess {
		// 	// publish rank withdraw
		// 	publishRankWithdraw(order)
		// }
	}
	resp, err := doHttpPost(reqUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 重复提现
func RepeatWithdrawNotify(req *data.RepeatWithdrawReq) ([]byte, error) {
	reqUrl := fmt.Sprintf("%s%s", gmUrl, "/withdraw/repeatWithdraw")
	strReq, _ := jsoniter.Marshal(req)
	resp, err := doHttpPost(reqUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit repeatWithdrawreq success, id:%s", req.OrderID)
	return resp, nil
}

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
