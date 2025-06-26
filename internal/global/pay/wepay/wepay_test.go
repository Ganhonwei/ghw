package wepay

import (
	"fmt"
	"testing"
)

func TestSign(t *testing.T) {
	// : : : : : : sign:5e3baace034dbee8abf17f44f3fe938a :
	dic := make(map[string]string)
	dic["amount"] = "100"
	dic["charge"] = "9"
	dic["orderAmount"] = "100"
	dic["orderNo"] = "243188500791300034"
	dic["otherData"] = ""
	dic["payStatus"] = "2"
	// dic["payTime"] = "0"
	dic["remark"] = "SBUN0014678"
	dic["reverse"] = "false"
	dic["tradeNo"] = "20240420132311322610"
	sign := PaySign(dic, "e7ecb56259eb498aa6d4dfda4fdb8b41")
	fmt.Println(sign) // 2178067e857930097c004e76b5fc8c70
}
