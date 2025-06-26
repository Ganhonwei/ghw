package atpay

import (
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

	return paramMap, nil
}

func (t *PayConfig) VerifySign(paramMap map[string]string) (bool, error) {
	outSign := paramMap["sign"]
	delete(paramMap, "sign")

	sign := PaySign(paramMap, t.MD5Key)

	return outSign == sign, nil
}

func (t *PayConfig) GetSuccessMsg() string {
	return "success"
}
