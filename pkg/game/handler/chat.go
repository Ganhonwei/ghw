package handler

import "goserver/gen/pb"

// ChatTextMsg 文本聊天消息
func ChatTextMsg(seat uint32, userid string, msg string) *pb.ChatTextRsp {
	return &pb.ChatTextRsp{
		Seat:    seat,
		Userid:  userid,
		Content: msg,
	}
}

// ChatVoiceMsg 语音聊天消息
// func ChatVoiceMsg(seat uint32, userid string, msg string) *pb.ChatVoiceRsp {
// 	return &pb.ChatVoiceRsp{
// 		Seat:    seat,
// 		Userid:  userid,
// 		Content: msg,
// 	}
// }

// n内容检测
func CheckContent(content string) bool {
	// 长度检查
	if len(content) > 200 {
		return false
	}
	// 内容 TODO
	return true
}
