package libs

import (
	"fmt"

	"github.com/bytedance/sonic"
)

const (
	CodeOK    = 200
	CodeFail  = 400
	CodeError = 500
)

type R map[string]any

// Response 响应体包装
type Response struct {
	Seq   int            `json:"seq,omitempty"`
	Code  int            `json:"code,omitempty"`
	Msg   string         `json:"msg,omitempty"`
	Data  any            `json:"data,omitempty"`
	Extra map[string]any `json:"extra,omitempty"`
}

func (resp *Response) Json() []byte {
	json, err := sonic.Marshal(resp)
	if err != nil {
		fmt.Println("unknown json error: ", err)
		return nil
	}
	return json
}

func (resp *Response) JsonString() string {
	return string(resp.Json())
}

func SeqResp(seq int, code int, msg string, data any) *Response {
	return &Response{
		Seq:  seq,
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func Resp(code int, msg string, data any) *Response {
	return SeqResp(0, code, msg, data)
}

func RSuccess(data any) *Response {
	return Resp(CodeOK, "", data)
}

func RFail(msg string) *Response {
	return Resp(CodeFail, msg, nil)
}

func RError(msg string) *Response {
	return Resp(CodeError, msg, nil)
}
