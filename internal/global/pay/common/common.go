package common

import (
	"fmt"
	"net/url"
	"strings"
)

// 分转元，不带小数点
func FenToYuanNoDecimal(fen uint32) string {
	return fmt.Sprintf("%d", fen/100)
}

// 分转元
func FenToYuan(fen uint32) string {
	return fmt.Sprintf("%.2f", float64(fen)/100)
}

// 元转分
func YuanToFen(yuan string) int {
	var amount float64
	fmt.Sscanf(yuan, "%f", &amount)
	return int(amount * 100)
}

// 分转分字符串
func FenToFenStr(fen uint32) string {
	return fmt.Sprintf("%d", fen)
}

// 分转分int
func FenToFenInt(fen string) int {
	var amount int
	fmt.Sscanf(fen, "%d", &amount)
	return amount
}

func IsValidURL(urlStr string) bool {
	// 1. 基本格式检查
	if urlStr == "" {
		return false
	}

	// 2. 规范化 URL
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	// 3. URL 解析
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	// 4. 检查必要组件
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	// 5. 发送 HEAD 请求检查可访问性
	// client := &http.Client{
	// 	Timeout: 5 * time.Second, // 设置超时
	// 	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	// 		if len(via) >= 10 { // 限制重定向次数
	// 			return http.ErrUseLastResponse
	// 		}
	// 		return nil
	// 	},
	// }

	// resp, err := client.Head(urlStr)
	// if err != nil {
	// 	return false
	// }
	// defer resp.Body.Close()

	// 6. 检查响应状态码
	// return resp.StatusCode >= 200 && resp.StatusCode < 400
	return true
}
