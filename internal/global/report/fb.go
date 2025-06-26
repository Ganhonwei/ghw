package report

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"goserver/pkg/utils"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
)

func reportFb(url string) (error, int, string) {
	// 创建请求
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err, -1, "NewRequest error"
	}

	// 设置Content-Type和Authorization头部
	// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Set("Authorization", "Bearer "+s2sCode)

	// 创建HTTP客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, -1, "Do error"
	}
	defer resp.Body.Close()

	return nil, resp.StatusCode, resp.Status
}

func handlerEventFBRegister(event *pb.FBEventRegister, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive FBEventRegister event: %v", event)
	delay = 1 * time.Minute

	e := data.GetFB(event.AppId)
	if e == nil || e.AppId == "" {
		ack = true
		return
	}

	// 获取fb参数
	val, err := myredis.Redis().Get(context.Background(), fmt.Sprintf("user:%s", event.Userid)).Result()
	if err != nil {
		if err == redis.Nil && time.Since(metadata.Timestamp) > time.Hour {
			data.ADReport2(event.Userid, event.AppId, "", "regist", metadata.Timestamp.Unix(), -1, "timeout")
			ack = true
		} else {
			ack = false
		}
		return
	}

	eventAdParam := &pb.EventAdParam{}
	err = eventAdParam.Unmarshal([]byte(val))
	if err != nil {
		ack = false
		return
	}

	phones := make([]string, 0)
	if eventAdParam.Phone != "" {
		// 散列化
		hashed := sha256.Sum256([]byte(eventAdParam.Phone))
		phones = append(phones, fmt.Sprintf("%x", hashed))
	}

	// 地理位置信息
	ip := strings.ReplaceAll(strings.ReplaceAll(eventAdParam.Ip, "[", ""), "]", "")
	addr := net.ParseIP(ip)
	record, err1 := ipClient.City(addr)

	eventDatas := make([]data.FBBaseEvent, 0)
	f := data.FBBaseEvent{
		EventName:      "CompleteRegistration",
		EventTime:      utils.String(utils.LocalTime().Unix()),
		ActionSource:   "website",
		EventSourceUrl: e.URL,
		UserData: data.FBUserData{
			Fbc:             eventAdParam.Fbc,
			Fbp:             eventAdParam.Fbp,
			ClientUserAgent: eventAdParam.UserAgent,
			ClientIpAddress: ip,
		},
	}
	if len(phones) > 0 {
		f.UserData.Phone = phones
	}
	if err1 == nil {
		// 国家
		isoCode := strings.ToLower(record.Country.IsoCode)
		hashed := sha256.Sum256([]byte(isoCode))
		f.UserData.Country = append(f.UserData.Country, fmt.Sprintf("%x", hashed))
		// 城市
		city := strings.ToLower(record.City.Names["en"])
		// 去空格
		city = strings.ReplaceAll(city, " ", "")
		hashed = sha256.Sum256([]byte(city))
		f.UserData.City = append(f.UserData.City, fmt.Sprintf("%x", hashed))
		// 邮编
		mailCode := strings.ToLower(record.Postal.Code)
		// 去空格
		mailCode = strings.ReplaceAll(mailCode, " ", "")
		// 去破折号
		mailCode = strings.ReplaceAll(mailCode, "-", "")
		hashed = sha256.Sum256([]byte(mailCode))
		f.UserData.MailCode = append(f.UserData.MailCode, fmt.Sprintf("%x", hashed))
	}

	eventDatas = append(eventDatas, f)
	datas, _ := json.Marshal(eventDatas)

	d := url.QueryEscape(string(datas))

	var fburl string
	if env == "dev" {
		fburl = fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?data=%s&access_token=%s&test_event_code=TEST83395", e.Pixel, d, e.AccessToken)
	} else {
		fburl = fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?data=%s&access_token=%s", e.Pixel, d, e.AccessToken)
	}

	err, status_code, status := reportFb(fburl)
	if err != nil {
		ack = false
		return
	}

	data.FBReport(event.Userid, event.AppId, "regist", metadata.Timestamp.Unix(), status_code, status, string(datas))
	ack = true
	return
}

// fb首次充值事件
func FBFirstDepositHandler(event *data.FBEvent, f data.FBBaseEvent) (int, string, error) {
	f.EventName = "FirstPurchase"

	eventDatas := []data.FBBaseEvent{f}
	datas, _ := json.Marshal(eventDatas)

	d := url.QueryEscape(string(datas))

	var fburl string
	if env == "dev" {
		fburl = fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?data=%s&access_token=%s&test_event_code=TEST83395", event.Pixel, d, event.AccessToken)
	} else {
		fburl = fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?data=%s&access_token=%s", event.Pixel, d, event.AccessToken)
	}

	err, status_code, status := reportFb(fburl)
	if err != nil {
		return 0, "", err
	}
	return status_code, status, nil
}

func handlerEventFBDeposit(event *pb.FBEventDeposit, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive FBEventDeposit event: %v", event)
	delay = 1 * time.Minute

	e := data.GetFB(event.AppId)
	if e == nil || e.AppId == "" {
		ack = true
		return
	}

	// 获取fb参数
	val, err := myredis.Redis().Get(context.Background(), fmt.Sprintf("user:%s", event.Userid)).Result()
	if err != nil {
		if err == redis.Nil && time.Since(metadata.Timestamp) > time.Hour {
			data.ADReport2(event.Userid, event.AppId, "", "deposit", metadata.Timestamp.Unix(), -1, "timeout")
			ack = true
		} else {
			ack = false
		}
		return
	}

	eventAdParam := &pb.EventAdParam{}
	err = eventAdParam.Unmarshal([]byte(val))
	if err != nil {
		ack = false
		return
	}

	phones := make([]string, 0)
	if eventAdParam.Phone != "" {
		// 散列化
		hashed := sha256.Sum256([]byte(eventAdParam.Phone))
		phones = append(phones, fmt.Sprintf("%x", hashed))
	}

	// 地理位置信息
	ip := strings.ReplaceAll(strings.ReplaceAll(eventAdParam.Ip, "[", ""), "]", "")
	addr := net.ParseIP(ip)
	record, err1 := ipClient.City(addr)

	eventDatas := make([]data.FBBaseEvent, 0)
	f := data.FBBaseEvent{
		EventName:      "Purchase",
		EventTime:      utils.String(utils.LocalTime().Unix()),
		ActionSource:   "website",
		EventSourceUrl: e.URL,
		UserData: data.FBUserData{
			Fbc:             eventAdParam.Fbc,
			Fbp:             eventAdParam.Fbp,
			ClientUserAgent: eventAdParam.UserAgent,
			ClientIpAddress: ip,
		},
		CustomData: data.FBCustomData{
			Currency: "INR",
			Value:    event.Amount,
		},
	}
	if len(phones) > 0 {
		f.UserData.Phone = phones
	}
	if err1 == nil {
		// 国家
		isoCode := strings.ToLower(record.Country.IsoCode)
		hashed := sha256.Sum256([]byte(isoCode))
		f.UserData.Country = append(f.UserData.Country, fmt.Sprintf("%x", hashed))
		// 城市
		city := strings.ToLower(record.City.Names["en"])
		// 去空格
		city = strings.ReplaceAll(city, " ", "")
		hashed = sha256.Sum256([]byte(city))
		f.UserData.City = append(f.UserData.City, fmt.Sprintf("%x", hashed))
		// 邮编
		mailCode := strings.ToLower(record.Postal.Code)
		// 去空格
		mailCode = strings.ReplaceAll(mailCode, " ", "")
		// 去破折号
		mailCode = strings.ReplaceAll(mailCode, "-", "")
		hashed = sha256.Sum256([]byte(mailCode))
		f.UserData.MailCode = append(f.UserData.MailCode, fmt.Sprintf("%x", hashed))
	}

	eventDatas = append(eventDatas, f)
	datas, _ := json.Marshal(eventDatas)

	d := url.QueryEscape(string(datas))

	var fburl string
	if env == "dev" {
		fburl = fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?data=%s&access_token=%s&test_event_code=TEST83395", e.Pixel, d, e.AccessToken)
	} else {
		fburl = fmt.Sprintf("https://graph.facebook.com/v19.0/%s/events?data=%s&access_token=%s", e.Pixel, d, e.AccessToken)
	}

	err, status_code, status := reportFb(fburl)
	if err != nil {
		ack = false
		return
	}

	data.FBReport(event.Userid, event.AppId, "deposit", metadata.Timestamp.Unix(), status_code, status, string(datas))

	// 首充
	if event.First {
		var status_code int
		var status string
		status_code, status, err = FBFirstDepositHandler(e, f)
		if err != nil {
			ack = false
			return
		}
		data.FBReport(event.Userid, event.AppId, "firstDeposit", metadata.Timestamp.Unix(), status_code, status, string(datas))
	}

	ack = true
	return
}
