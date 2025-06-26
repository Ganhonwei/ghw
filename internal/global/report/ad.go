package report

import (
	"context"
	"fmt"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/myredis"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
)

func reportAd(formData url.Values, s2sCode string) (error, int, string) {
	// 创建请求
	req, err := http.NewRequest("POST", "https://s2s.adjust.com/event", strings.NewReader(formData.Encode()))
	if err != nil {
		return err, -1, "NewRequest error"
	}

	// 设置Content-Type和Authorization头部
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer "+s2sCode)

	// 创建HTTP客户端并发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err, -1, "Do error"
	}
	defer resp.Body.Close()

	return nil, resp.StatusCode, resp.Status
}

func handlerEventRegister(event *pb.EventRegister, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive event: %v", event)
	delay = 1 * time.Minute

	//获取ad参数
	val, err := myredis.Redis().Get(context.Background(), fmt.Sprintf("user:%s", event.Userid)).Result()
	if err != nil {
		if err == redis.Nil && time.Since(metadata.Timestamp) > time.Hour {
			data.ADReport2(event.Userid, event.BundleId, "", "register", metadata.Timestamp.Unix(), -1, "timeout")
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

	formData := url.Values{
		"adid":        {eventAdParam.AdId},
		"event_token": {eventAdParam.AdEventCodeRegister},
		"app_token":   {eventAdParam.AdKey},
		"s2s":         {"1"},
		"user_agent":  {eventAdParam.UserAgent},
		"ip_address":  {eventAdParam.Ip},
	}

	//上报ad
	err, status_code, status := reportAd(formData, eventAdParam.AdS2SCode)
	if err != nil {
		ack = false
		return
	}

	data.ADReport2(event.Userid, event.BundleId, eventAdParam.AdId, "register", metadata.Timestamp.Unix(), status_code, status)
	ack = true
	return
}

func handlerEventLogin(event *pb.EventLogin, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive event: %v", event)
	delay = 1 * time.Minute

	//获取ad参数
	val, err := myredis.Redis().Get(context.Background(), fmt.Sprintf("user:%s", event.Userid)).Result()
	if err != nil {
		if err == redis.Nil && time.Since(metadata.Timestamp) > time.Hour {
			data.ADReport2(event.Userid, event.BundleId, "", "login", metadata.Timestamp.Unix(), -1, "timeout")
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

	formData := url.Values{
		"adid":        {eventAdParam.AdId},
		"event_token": {eventAdParam.AdEventCodeLogin},
		"app_token":   {eventAdParam.AdKey},
		"s2s":         {"1"},
		"user_agent":  {eventAdParam.UserAgent},
		"ip_address":  {eventAdParam.Ip},
	}

	//上报ad
	err, status_code, status := reportAd(formData, eventAdParam.AdS2SCode)
	if err != nil {
		ack = false
		return
	}

	data.ADReport2(event.Userid, event.BundleId, eventAdParam.AdId, "login", metadata.Timestamp.Unix(), status_code, status)
	ack = true
	return
}

func handlerEventDeposit(event *pb.EventDeposit, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive event: %v", event)
	delay = 1 * time.Minute

	//获取ad参数
	val, err := myredis.Redis().Get(context.Background(), fmt.Sprintf("user:%s", event.Userid)).Result()
	if err != nil {
		if err == redis.Nil && time.Since(metadata.Timestamp) > time.Hour {
			data.ADReport2(event.Userid, event.BundleId, "", "deposit", metadata.Timestamp.Unix(), -1, "timeout")
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

	formData := url.Values{
		"adid":        {eventAdParam.AdId},
		"event_token": {eventAdParam.AdEventCodeDeposit},
		"app_token":   {eventAdParam.AdKey},
		"s2s":         {"1"},
		"revenue":     {fmt.Sprintf("%f", event.Amount)},
		"currency":    {"INR"},
		"user_agent":  {eventAdParam.UserAgent},
		"ip_address":  {eventAdParam.Ip},
	}

	//上报ad
	err, status_code, status := reportAd(formData, eventAdParam.AdS2SCode)
	if err != nil {
		ack = false
		return
	}

	data.ADReport2(event.Userid, event.BundleId, eventAdParam.AdId, "deposit", metadata.Timestamp.Unix(), status_code, status)

	// 上报首充
	if event.First && eventAdParam.AdEventCodeFirstDeposit != "" {
		formData := url.Values{
			"adid":        {eventAdParam.AdId},
			"event_token": {eventAdParam.AdEventCodeFirstDeposit},
			"app_token":   {eventAdParam.AdKey},
			"s2s":         {"1"},
			"revenue":     {fmt.Sprintf("%f", event.Amount)},
			"currency":    {"INR"},
			"user_agent":  {eventAdParam.UserAgent},
			"ip_address":  {eventAdParam.Ip},
		}

		//上报ad
		var status_code int
		var status string
		err, status_code, status = reportAd(formData, eventAdParam.AdS2SCode)
		if err != nil {
			ack = false
			return
		}

		data.ADReport2(event.Userid, event.BundleId, eventAdParam.AdId, "firstDeposit", metadata.Timestamp.Unix(), status_code, status)
	}

	ack = true
	return
}

func handlerEventAdParam(event *pb.EventAdParam, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive event: %v", event)
	delay = 1 * time.Minute

	var body []byte
	body, err = event.Marshal()
	if err != nil {
		ack = false
		return
	}

	//保存ad参数
	err = myredis.Redis().Set(context.Background(), fmt.Sprintf("user:%s", event.Userid), body, 0).Err()
	if err != nil {
		ack = false
		return
	}

	ack = true
	return
}
