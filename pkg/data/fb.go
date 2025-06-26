package data

import "github.com/globalsign/mgo/bson"

type FBEvent struct {
	AppId       string `json:"appId" bson:"_id"`
	Pixel       string `json:"pixel" bson:"pixel"`
	AccessToken string `json:"accessToken" bson:"access_token"`
	URL         string `json:"url" bson:"event_source_url"`
}

type FBBaseEvent struct {
	EventName      string       `json:"event_name" `
	EventTime      string       `json:"event_time" `
	ActionSource   string       `json:"action_source"`
	EventSourceUrl string       `json:"event_source_url"`
	UserData       FBUserData   `json:"user_data"`
	CustomData     FBCustomData `json:"custom_data,omitempty"`
}

type FBUserData struct {
	Fbc             string   `json:"fbc,omitempty"`
	Fbp             string   `json:"fbp,omitempty"`
	Phone           []string `json:"phone,omitempty"`
	ClientUserAgent string   `json:"client_user_agent,omitempty"`
	ClientIpAddress string   `json:"client_ip_address,omitempty"`
	Country         []string `json:"country,omitempty"`
	City            []string `json:"ct,omitempty"`
	MailCode        []string `json:"zp,omitempty"`
}

type FBCustomData struct {
	Currency string  `json:"currency,omitempty"`
	Value    float64 `json:"value,omitempty"`
}

func GetFB(appid string) *FBEvent {
	e := new(FBEvent)
	Get(FBReports, appid, &e)
	return e
}

func FBReport(userid, appid, etype string, etime int64, status_code int, status, param string) {
	var code int
	if status_code == 200 {
		code = 0
	} else {
		code = -1
	}

	record := &LogFBReport{
		Id:         bson.NewObjectId().String(),
		AppId:      appid,
		Userid:     userid,
		Type:       etype,
		Code:       code,
		Etime:      etime,
		Ctime:      bson.Now().Unix(),
		StatusCode: status_code,
		Status:     status,
		ReqParam:   param,
	}
	record.Save()
}
