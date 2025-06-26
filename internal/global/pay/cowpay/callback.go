package cowpay

import (
	"goserver/pkg/data"

	"github.com/valyala/fasthttp"
)

func (t *PayConfig) CheckMethod(ctx *fasthttp.RequestCtx) bool {
	return ctx.IsPost()
}

func (t *PayConfig) ParseParamMap(ctx *fasthttp.RequestCtx) (map[string]string, error) {
	body := ctx.PostBody()

	paramMap, err := ParsePayResult(body)
	if err != nil {
		return nil, err
	}

	transdata := data.ToUrlDecode(paramMap["transdata"])
	sign := data.ToUrlDecode(paramMap["sign"])

	paramMap2, err := ParsePayResult([]byte(transdata))
	if err != nil {
		return nil, err
	}

	paramMap2["sign"] = sign

	return paramMap2, nil
}

func (t *PayConfig) VerifySign(paramMap map[string]string) (bool, error) {
	outSign := paramMap["sign"]
	delete(paramMap, "sign")

	sign := PaySign(paramMap, t.MD5Key)
	sign = data.ToUrlEncode(data.ToUpper(sign))

	return outSign == sign, nil
}

func (t *PayConfig) GetSuccessMsg() string {
	return "success"
}
