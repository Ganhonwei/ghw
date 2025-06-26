package args

// 客服回复消息
type CustomerReplayArgs struct {
	Userid  string `json:"userid"`
	Ctype   int32  `json:"ctype"`
	Content string `json:"content"`
}

// 用户会话信息
type UserSessionsArgs struct {
	Stime    string `json:"stime"`
	Etime    string `json:"etime"`
	Userid   string `json:"userid"`
	Customer string `json:"customer"`
	Status   int32  `json:"status"`
}

// 客服界面2全部消息
type AllSessionsArgs struct {
	Stime        string `json:"stime"`
	Etime        string `json:"etime"`
	Userid       string `json:"userid"`
	QuestionType int32  `json:"questionType"`
	Resolved     int32  `json:"resolved"` // 0全部 1未解决 2已解决
}

// 客服表现
type CustomerExpStatsArgs struct {
	Stime    string `json:"stime"`
	Etime    string `json:"etime"`
	Userid   string `json:"userid"`
	Customer string `json:"customer"`
}
