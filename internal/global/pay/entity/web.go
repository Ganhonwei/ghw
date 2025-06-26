package entity

type WebResponse struct {
	Code   int    `json:"code"` // 错误码    200:success
	ErrMsg string `json:"msg"`  // 错误信息
	Body   []byte `json:"body"` // 响应体
}

type WebQueryBalence struct {
	ChannelId uint32 `json:"channel_id"` // 渠道id
	Balence   int64  `json:"balence"`    // 可用余额
}
