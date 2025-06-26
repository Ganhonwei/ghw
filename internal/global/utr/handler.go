package utr

import (
	"context"
	"encoding/json"
	"errors"
	"goserver/gen/pb"
	"goserver/pkg/data"
	"goserver/pkg/data/mq"
	"goserver/pkg/glog"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/nats-io/nats.go/jetstream"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
)

var limiter *rate.Limiter

func init() {
	limiter = rate.NewLimiter(rate.Every(time.Minute/15), 1)
}

// 定义响应结构体
type PaymentResponse struct {
	RefNo     string `json:"refNo"`
	RefType   string `json:"refType"`
	PayTime   string `json:"payTime"`
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

// 验证函数
func (r *PaymentResponse) Validate() error {
	if r.RefNo == "" {
		return errors.New("missing reference number")
	}

	if r.RefType != "UTR" && r.RefType != "UPI" {
		return errors.New("invalid reference type")
	}

	if r.PayTime == "" {
		return errors.New("invalid payment time")
	}

	if r.Amount == "" {
		return errors.New("invalid amount format")
	}

	if r.Recipient == "" {
		return errors.New("missing recipient")
	}

	return nil
}

func handlerEventUploadUtr(event *pb.UploadUtrReq, metadata *jetstream.MsgMetadata) (ack bool, delay time.Duration, err error) {
	glog.Debugf("receive event: %v", event)
	delay = 1 * time.Minute

	// 尝试获取令牌
	if !limiter.Allow() {
		glog.Error("rate limit exceeded, retry later")
		ack = false
		delay = 4 * time.Second
		return
	}

	c := &http.Client{Transport: &ProxyRoundTripper{
		APIKey:   "AIzaSyClOaDoRRRKo0COz_G1Q0GYIqvgozbWGnc",
		ProxyURL: proxyUrl,
	}}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithHTTPClient(c), option.WithAPIKey("AIzaSyClOaDoRRRKo0COz_G1Q0GYIqvgozbWGnc"))

	// client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyClOaDoRRRKo0COz_G1Q0GYIqvgozbWGnc"))

	if err != nil {
		glog.Error(err)
		ack = false
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")

	// url := "https://bucket-indiagame-adjust.s3.ap-southeast-1.amazonaws.com/utr/photo_2025-04-23_14-24-45.jpg"
	// url := "https://bucket-indiagame-adjust.s3.ap-southeast-1.amazonaws.com/utr/photo_2025-04-24_15-54-34.jpg"

	// Download the image.
	imageResp, err := http.Get(event.FileName)
	if err != nil {
		glog.Error(err)
		ack = false
		return
	}
	defer imageResp.Body.Close()

	imageBytes, err := io.ReadAll(imageResp.Body)
	if err != nil {
		glog.Error(err)
		ack = false
		return
	}

	// Create the request.
	req := []genai.Part{
		genai.ImageData("jpeg", imageBytes),
		genai.Text(`请分析图片中的支付信息，严格按照以下JSON格式返回：
	{
		"refNo": "",        // UTR号码或UPI Ref No
		"refType": "",      // 标识是"UTR"还是"UPI"
		"payTime": "",      // 支付时间
		"amount": "",       // 支付金额
		"recipient": "",    // 收款方名称
	}
	注意：
	1. 必须是合法的JSON格式
	2. 只返回JSON数据，不要包含其他文字
	3. 如果无法识别某个字段，对应值返回空字符串""
	4. refNo字段优先识别UTR号码，如果是UPI Ref No也可以填入
	5. refType必须是"UTR"或"UPI"其中之一
	6. amount必须只包含数字，不要包含货币符号`),
	}

	// Generate content.
	resp, err := model.GenerateContent(ctx, req...)
	if err != nil {
		glog.Error(err)
		ack = false
		return
	}

	var response PaymentResponse
	// Handle the response of generated text.
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
		if !ok {
			glog.Error("响应格式不是预期的文本类型")
			ack = false
			return
		}

		re := regexp.MustCompile("```json\\s*([\\s\\S]*?)```")
		matches := re.FindStringSubmatch(string(text))
		if len(matches) <= 1 {
			glog.Error("解析JSON失败: 没有找到JSON数据")
			ack = false
			return
		}

		if err = json.Unmarshal([]byte(matches[1]), &response); err != nil {
			glog.Error("解析JSON失败:", err)
			ack = false
			return
		}

		// 验证和处理数据
		if err = response.Validate(); err != nil {
			glog.Error("数据验证失败:", err)
			ack = true //无法解析，不再处理
			// return
		}

		glog.Debugf("解析后的数据: %+v", response)
	} else {
		glog.Error("Gemini 未返回有效响应")
		ack = false
		return
	}

	if ack { // 无法解析，返回失败
		rsp := &pb.UploadUtrRsp{
			Uid:      event.Uid,
			OrderId:  event.OrderId,
			FileName: event.FileName,
			Status:   int32(data.UtrStatusFailed),
		}

		err = mq.NatsPublish(mq.TopicUtrUploadResult, rsp)
		if err != nil {
			glog.Error(err)
			ack = false //消息发送失败，不确认
			return
		}
	} else { //解析成功
		rsp := &pb.UploadUtrRsp{
			Uid:       event.Uid,
			OrderId:   event.OrderId,
			FileName:  event.FileName,
			Status:    int32(data.UtrStatusSuccess),
			RefNo:     response.RefNo,
			RefType:   response.RefType,
			PayTime:   response.PayTime,
			Amount:    response.Amount,
			Recipient: response.Recipient,
		}

		err = mq.NatsPublish(mq.TopicUtrUploadResult, rsp)
		if err != nil {
			glog.Error(err)
			ack = false //消息发送失败，不确认
			return
		}
		ack = true //消息发送成功，确认
		err = mq.NatsPublish(mq.TopicFillOrder, &pb.UtrFillOrder{
			ImageUrl:  event.FileName,
			OrderId:   event.OrderId,
			ChannelId: event.ChannelId,
		})
		if err != nil {
			glog.Error(err)
			return
		}
	}

	return
}
