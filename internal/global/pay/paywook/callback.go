package paywook

import (
	"fmt"
	"goserver/pkg/utils"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

func (t *PayConfig) CheckMethod(ctx *fasthttp.RequestCtx) bool {
	return ctx.IsPost()
}

func (t *PayConfig) ParseParamMap(ctx *fasthttp.RequestCtx) (map[string]string, error) {
	// 获取 Content-Type
	contentType := string(ctx.Request.Header.ContentType())

	// 创建返回的 map
	paramMap := make(map[string]string)

	// 处理 multipart/form-data
	if strings.Contains(contentType, "multipart/form-data") {
		// 解析 multipart 表单
		form, err := ctx.MultipartForm()
		if err != nil {
			return nil, err
		}

		// 遍历所有表单字段
		for key, values := range form.Value {
			if len(values) > 0 {
				paramMap[key] = values[0]
			}
		}

		return paramMap, nil
	}

	body := ctx.PostBody()
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}

	return ValuesToMap(values), nil
}

func (t *PayConfig) VerifySign(paramMap map[string]string) (bool, error) {
	outSign := paramMap["sign"]
	// delete(paramMap, "sign")

	// sign := PaySign(paramMap, t.MD5Key)

	siteid := paramMap["siteid"]
	orderid := paramMap["orderid"]
	currency := paramMap["currency"]
	amount := paramMap["amount"]
	invinceno := paramMap["invinceno"]
	verified := paramMap["verified"]

	str := fmt.Sprintf("%s%s%s%s%s%s%s", siteid, orderid, currency, amount, invinceno, verified, t.MD5Key)
	sign := utils.Md5(str)

	return outSign == sign, nil
}

func (t *PayConfig) GetSuccessMsg() string {
	return "success"
}
