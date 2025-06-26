package wepay2

import (
	"net/url"

	"github.com/valyala/fasthttp"
)

func (t *PayConfig) CheckMethod(ctx *fasthttp.RequestCtx) bool {
	return ctx.IsPost()
}

func (t *PayConfig) ParseParamMap(ctx *fasthttp.RequestCtx) (map[string]string, error) {
	body := ctx.PostBody()

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	paramMap := ValuesToMap(values)

	return paramMap, nil
}

func (t *PayConfig) VerifySign(paramMap map[string]string) (bool, error) {
	outSign := paramMap["sign"]
	delete(paramMap, "sign")
	delete(paramMap, "signType")
	delete(paramMap, "utr")
	delete(paramMap, "merRetMsg")

	sign := PaySign(paramMap, t.PayMD5Key)
	if outSign == sign {
		return true, nil
	}

	delete(paramMap, "message")

	sign = PaySign(paramMap, t.WithdrawMD5Key)
	if outSign == sign {
		return true, nil
	}

	return false, nil
}

func (t *PayConfig) GetSuccessMsg() string {
	return "success"
}
