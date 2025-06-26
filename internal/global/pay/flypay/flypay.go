package flypay

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"goserver/internal/global/pay/entity"
	"goserver/pkg/data"
	"goserver/pkg/glog"
	"goserver/pkg/utils"
	"io"
	"math/big"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var Config *FlypayConfig

const (
	pem_begin = "-----BEGIN PRIVATE KEY-----\n"
	pem_end   = "\n-----END PRIVATE KEY-----"
	pub_begin = "-----BEGIN PUBLIC KEY-----\n"
	pub_end   = "\n-----END PUBLIC KEY-----"
)

// Submit 下单
func (t *FlypayConfig) Submit(order *entity.PayOrder) ([]byte, error) {
	body := t.compose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	//fmt.Printf("body %s, order %#v\n", body, order)
	resp, err := doHttpForm(t.PayOrderUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Submit order success, id:%s", order.OrderID)
	return resp, nil
}

// 提现下单
func (t *FlypayConfig) WithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	body := t.WithdrawCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	resp, err := doHttpForm(t.WithDrawUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("WithdrawSubmit order success, id:%s", order.OrderID)
	return resp, nil
}

// InspectSubmit 核查
func (t *FlypayConfig) InspectSubmit(order *entity.PayOrder) ([]byte, error) {
	body := t.inspectCompose(order) //打包post数据
	// strReq := pay.ToQueryString(body)
	strReq, _ := jsoniter.Marshal(body)
	order.RequestParam = string(strReq)
	//fmt.Printf("body %s, order %#v\n", body, order)
	resp, err := doHttpForm(t.InspectOrderUrl, strReq)
	//fmt.Printf("resp %s, err %v", string(resp), err)
	if err != nil {
		return nil, err
	}
	glog.Infof("Inspect order success, id:%s", order.OrderID)
	return resp, nil
}

// 组装参数
func (c *FlypayConfig) compose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)
	dictionary["mer_no"] = c.MachId
	dictionary["mer_order_no"] = order.OrderID
	dictionary["pname"] = "game player"
	dictionary["pemail"] = order.Email
	dictionary["phone"] = order.Mobile
	dictionary["order_amount"] = strconv.Itoa(int(order.Amount) / 100)
	dictionary["country_code"] = "IND"
	dictionary["cyy_no"] = "INR"
	dictionary["pay_type"] = "UPI"
	dictionary["notify_url"] = c.PayNotifyUrl

	// 签名
	strSign := PaySign(c.MachId, order.OrderID, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 组装参数
func (c *FlypayConfig) inspectCompose(order *entity.PayOrder) map[string]string {
	dictionary := make(map[string]string)
	dictionary["mer_no"] = c.MachId
	dictionary["mer_order_no"] = order.OrderID

	// 签名
	strSign := PaySign(c.MachId, order.OrderID, c.Md5Key)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

// 提现组装参数
func (t *FlypayConfig) WithdrawCompose(order *entity.WithdrawOrder) map[string]string {
	dictionary := make(map[string]string)
	dictionary["mer_no"] = t.MachId
	dictionary["mer_order_no"] = order.OrderID
	dictionary["order_amount"] = strconv.Itoa(int(order.Amount) / 100)
	dictionary["pay_type"] = "BANK"
	dictionary["cyy_no"] = "INR"
	dictionary["acc_name"] = order.RealName
	dictionary["acc_no"] = order.BankNumber
	dictionary["province"] = order.IFSC
	dictionary["phone"] = order.Mobile
	dictionary["email"] = order.Email
	// dictionary["summary"] = ""
	dictionary["notifyurl"] = t.WithDrawNotifyUrl

	// 签名
	// strSign := WithdrawSign(order.OrderID, timestamp, t.Md5Key)
	strdic, _ := json.Marshal(dictionary)
	strSign := rsaSign(strdic, t.PrivateKey, crypto.SHA256)
	// dictionary["sign"] = pay.ToUpper(strSign)
	dictionary["sign"] = strSign
	return dictionary
}

func PaySign(mId, orderId, md5key string) string {
	result := mId + orderId + md5key
	return utils.Md5(result)
}

func rsaSign(content []byte, key string, hash crypto.Hash) string {
	shaNew := hash.New()
	shaNew.Write([]byte(content))
	hashed := shaNew.Sum(nil)
	prikey, err := parsePrivateKey(Config.PrivateKey)
	if err != nil {
		glog.Errorf("get prikey fail,err:%s", err)
		return ""
	}
	sign, err1 := rsa.SignPKCS1v15(rand.Reader, prikey, hash, hashed)
	if err1 != nil {
		glog.Error("rsa fail,err:%s", err1)
		return ""
	}
	s := base64.StdEncoding.EncodeToString(sign)
	// b, _ := base64.StdEncoding.DecodeString(s)
	// shaNew = hash.New()
	// shaNew.Write([]byte(content))
	// hashed = shaNew.Sum(nil)
	// pubkey, _ := parsePublickKey("MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAMc8H3hgSEiSCoSbWOYsUzd34cqdFzE3HJza7lr3Zofo2IrdLsN4r0lpMsLTCgHF6NxC3akUzSvkWMnRtsaU+Y0CAwEAAQ==")
	// err2 := rsa.VerifyPKCS1v15(pubkey, hash, hashed[:], b)
	// if err2 != nil {
	// 	return s
	// }
	return s
}

// 生成私钥对象
func parsePrivateKey(key string) (*rsa.PrivateKey, error) {
	if !strings.HasPrefix(key, pem_begin) {
		key = pem_begin + key
	}
	if !strings.HasSuffix(key, pem_end) {
		key = key + pem_end
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("key is error")
	}
	// 解析DER编码的私钥，生成私钥对象
	prikey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return prikey.(*rsa.PrivateKey), nil
}

// 生成公钥对象
func parsePublicKey(key string) (*rsa.PublicKey, error) {
	if !strings.HasPrefix(key, pub_begin) {
		key = pub_begin + key
	}
	if !strings.HasSuffix(key, pub_end) {
		key = key + pub_end
	}
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		return nil, errors.New("key is error")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to parse RSA public key")
	}
	return rsaPubKey, nil
}

func (c *FlypayConfig) VerifySing(sign string, content []byte) bool {
	key := c.ServerPublicKey
	if !strings.HasPrefix(key, pub_begin) {
		key = pub_begin + key
	}
	if !strings.HasSuffix(key, pub_end) {
		key = key + pub_end
	}
	block, _ := pem.Decode([]byte(key))
	public, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Printf("加载公钥错误 %+v\n", err)
		return false
	}
	decodeBytes, _ := base64.StdEncoding.DecodeString(sign)
	var result string
	// 因为加密解密用rsa有长度限值，因此这个地方就看具体限值的多少了
	for len(decodeBytes) > 0 {
		var decodePart []byte
		if len(decodeBytes) < 128 {
			decodePart = decodeBytes
		} else {
			decodePart = decodeBytes[:128]
		}
		plain := RSA_public_decrypt(public.(*rsa.PublicKey), decodePart)
		result += string(plain)
		decodeBytes = decodeBytes[len(decodePart):]
	}
	if result != "" {
		tradeResult := WithdrawCallback{}
		err := jsoniter.UnmarshalFromString(result, &tradeResult)
		if err != nil {
			return false
		}
	}
	fmt.Printf("%+v\n", result)
	return true
}

func RSA_public_decrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}

// application/x-www-form-urlencoded PSOT请求
func doHttpForm(targetUrl string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", targetUrl, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}
	req.Header.Add("Content-type", "application/json")

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   0,
			KeepAlive: 0,
		}).Dial,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("err:%s", err)
		return []byte(""), err
	}

	return respData, nil
}

func (c *FlypayConfig) SubmitResponse(body []byte, order *entity.PayOrder) (string, error) {
	result := new(OrderRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 1 { // 没成功TODO
		return "", fmt.Errorf("code:%d, OrderNo:%s", result.Code, result.Mer_order_no)
	}
	return result.Pay_url, nil
	// return "", nil
}

func (c *FlypayConfig) ParsePayResult(body []byte) (map[string]string, error) {
	tradeResult := RechargeCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["code"] = fmt.Sprint(tradeResult.Code)
	reslut["msg"] = tradeResult.Msg
	reslut["order_status"] = tradeResult.Order_status
	reslut["mer_no"] = tradeResult.Mer_no
	reslut["mer_order_no"] = tradeResult.Mer_order_no
	reslut["order_amount"] = tradeResult.Order_amount
	reslut["pay_amount"] = tradeResult.Pay_amount
	reslut["order_no"] = tradeResult.Order_no
	reslut["order_time"] = fmt.Sprint(tradeResult.Order_time)
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *FlypayConfig) ParseWithdrawResult(body []byte) (map[string]string, error) {
	tradeResult := WithdrawCallback{}
	err := jsoniter.Unmarshal(body, &tradeResult)
	if err != nil {
		return nil, err
	}
	reslut := make(map[string]string)
	reslut["code"] = fmt.Sprint(tradeResult.Code)
	reslut["msg"] = tradeResult.Msg
	reslut["order_status"] = tradeResult.Order_status
	reslut["mer_no"] = tradeResult.Mer_no
	reslut["mer_order_no"] = tradeResult.Mer_order_no
	reslut["order_amount"] = tradeResult.Order_amount
	reslut["order_no"] = tradeResult.Order_no
	reslut["order_time"] = fmt.Sprint(tradeResult.Order_time)
	reslut["sign"] = tradeResult.Sign
	return reslut, nil
}

func (c *FlypayConfig) WithdrawSubmitResponse(body []byte, order *entity.WithdrawOrder) (bool, string) {
	result := new(WithdrawRespond)
	jsoniter.Unmarshal(body, result)
	if result.Code != 1 { // 没成功TODO
		glog.Errorf("[flypay] withdraw fail submit, code:%s, msg:%s", result.Code, result.Msg)
		order.OrderStatus = data.OrderFail
		return false, fmt.Sprint(result.Msg)
	}
	order.OrderStatus = data.OrderSuccess
	return true, fmt.Sprint(result.Code)
}

// 核单响应
func (c *FlypayConfig) InspectResponse(body []byte, order *entity.PayOrder) error {
	result := new(InspectResponse)
	jsoniter.Unmarshal(body, result)
	if result.Code != 1 { // 没成功TODO
		return fmt.Errorf("code:%d, message:%s", result.Code, result.Msg)
	}
	paramMap := make(map[string]string)
	paramMap["code"] = utils.String(result.Code)
	paramMap["msg"] = result.Msg
	paramMap["mer_no"] = result.MerNo
	paramMap["mer_order_no"] = result.MerOrderNo
	paramMap["order_number"] = result.OrderNumber
	paramMap["pay_amount"] = result.PayAmount
	paramMap["order_status"] = result.OrderStatus
	body, err := json.Marshal(paramMap)
	if err != nil {
		return fmt.Errorf("[flypay] marshal fail")
	}
	// 验签
	if !c.VerifySing(result.Sign, body) {
		return fmt.Errorf("[flypay] verify sign fail")
	}
	return nil
}

// 核单响应
func (c *FlypayConfig) InspectWdResponse(body []byte, order *entity.WithdrawOrder) int {
	// result := new(InspectResponse)
	// jsoniter.Unmarshal(body, result)
	// if result.Code != 1 { // 没成功TODO
	// 	return 3
	// }
	// paramMap := make(map[string]string)
	// paramMap["code"] = utils.String(result.Code)
	// paramMap["msg"] = result.Msg
	// paramMap["mer_no"] = result.MerNo
	// paramMap["mer_order_no"] = result.MerOrderNo
	// paramMap["order_number"] = result.OrderNumber
	// paramMap["pay_amount"] = result.PayAmount
	// paramMap["order_status"] = result.OrderStatus
	// body, err := json.Marshal(paramMap)
	// if err != nil {
	// 	return 3
	// }
	// // 验签
	// if !c.VerifySing(result.Sign, body) {
	// 	return 3
	// }
	return 3
}

// 提现查单接口
func (t *FlypayConfig) InspectWithdrawSubmit(order *entity.WithdrawOrder) ([]byte, error) {
	// body := t.InspectWdCompose(order) //打包post数据
	// strReq := ToQueryString(body)
	// resp, err := doHttpForm(t.InspectWithdrawUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("InspectWithdrawSubmit order success, id:%s", order.OrderID)
	// return resp, nil
	return nil, fmt.Errorf("fail")
}

// 余额查询
func (c *FlypayConfig) BalanceQuery() int64 {
	// param := map[string]any{
	// 	"merchantNo": c.Appid,
	// 	"area":       "印度",
	// }
	// strReq := ToQueryString(param)
	// resp, err := doHttpForm(c.QueryUrl, strReq)
	// if err != nil {
	// 	return nil, err
	// }
	// glog.Infof("BalanceQuery order success, id:%s", order.OrderID)
	// return resp, nil
	return 0
}
