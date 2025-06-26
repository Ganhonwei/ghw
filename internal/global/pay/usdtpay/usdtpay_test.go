package usdtpay

import (
	"fmt"
	"testing"
)

func TestPaySign(t *testing.T) {
	md5key := "epusdt_password_xasddawqe"

	// "order_id": "20220201030210321",
	// "amount": 42,
	// "notify_url": "http://example.com/notify",
	// "redirect_url": "http://example.com/redirect",
	// "signature": "1cd4b52df5587cfb1968b0c0c6e156cd"

	dic := make(map[string]any)
	dic["order_id"] = "252036883125633026"
	dic["amount"] = 200
	dic["notify_url"] = "http://192.168.0.112:16080/api/pay/notify/usdt"
	// dic["redirect_url"] = "http://example.com/redirect"
	sign := PaySign(dic, md5key)
	fmt.Println(sign) // 2178067e857930097c004e76b5fc8c70
}
