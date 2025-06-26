package data

type Adjust struct {
	AppToken        string `json:"app_token" bson:"app_token"`
	AppVersion      string `json:"app_version" bson:"app_version"`
	AppId           string `json:"app_id" bson:"app_id"`
	Environment     string `json:"environment" bson:"environment"`
	TrackerName     string `json:"tracker_name" bson:"tracker_name"`
	LastTrackerName string `json:"last_tracker_name" bson:"last_tracker_name"`
	NetworkName     string `json:"network_name" bson:"network_name"`
	CampaignName    string `json:"campaign_name" bson:"campaign_name"`
	AdgroupName     string `json:"adgroup_name" bson:"adgroup_name"`
	CreativeName    string `json:"creative_name" bson:"creative_name"`
	InstalledAt     string `json:"installed_at" bson:"installed_at"`
	ReinstalledAt   string `json:"reinstalled_at" bson:"reinstalled_at"`
	UninstalledAt   string `json:"uninstalled_at" bson:"uninstalled_at"`
	Adid            string `json:"adid" bson:"adid"`
	GpsAdid         string `json:"gps_adid" bson:"gps_adid"`
	AndroidId       string `json:"android_id" bson:"android_id"`
	SdkVersion      string `json:"sdk_version" bson:"sdk_version"`
	Language        string `json:"language" bson:"language"`
	City            string `json:"city" bson:"city"`
	DeviceType      string `json:"device_type" bson:"device_type"`
	DeviceName      string `json:"device_name" bson:"device_name"`
	IpAddress       string `json:"ip_address" bson:"ip_address"`
	ProxyIpAddress  string `json:"proxy_ip_address" bson:"proxy_ip_address"`
	UserAgent       string `json:"user_agent" bson:"user_agent"`
}

type AdjustEvent struct {
	AppId        string `json:"appId" bson:"_id"`
	AdKey        string `json:"adKey" bson:"ad_key"`
	AdS2sCode    string `json:"ads2sCode" bson:"ad_s2s_code"`
	AdEventCode1 string `json:"adEventCode1" bson:"ad_event_code1"`
	AdEventCode2 string `json:"adEventCode2" bson:"ad_event_code2"`
	AdEventCode3 string `json:"adEventCode3" bson:"ad_event_code3"`
	AdEventCode4 string `json:"adEventCode4" bson:"ad_event_code4"`
	AdEventCode5 string `json:"adEventCode5" bson:"ad_event_code5"`
	Remark       string `json:"remark" bson:"remark"`
}

func (t *Adjust) Save() bool {
	return Insert(AdjustCallback, t)
}
