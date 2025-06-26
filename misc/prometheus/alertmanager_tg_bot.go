package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	tele "gopkg.in/telebot.v3"
)

var (
	TgBot *TelegramBot
)

func main() {
	addr := ":9095"
	TgBot = NewTgBot("6542067431:AAHZsTjUz2VHqm-VRYkaqazI3rVMV5KL7Jg", "-4108600082")
	err := TgBot.Init()
	if err != nil {
		panic(err)
	}
	// TgBot.Send("111")

	httpLogger := log.New(os.Stdout, "[http] ", log.LstdFlags)
	serveMux := NewHttpServeMux(httpLogger)
	httpLogger.Printf("Http listen: http://%s\n", addr)
	err = http.ListenAndServe(addr, serveMux)
	if err != nil {
		log.Fatalln(err)
	}
}

func NewHttpServeMux(httpLogger *log.Logger) *http.ServeMux {
	m := http.NewServeMux()

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var err error
		defer func() {
			r.Body.Close()
			if err != nil {
				httpLogger.Println("error: ", err)
				w.WriteHeader(http.StatusBadRequest)
			}
		}()

		var message Message
		decoder := json.NewDecoder(r.Body)

		// var m = make(map[string]any)
		// err = decoder.Decode(&m)
		// if err != nil {
		// 	return
		// }
		// b, _ := json.Marshal(m)
		// httpLogger.Printf("body: %+v\n", string(b))
		err = decoder.Decode(&message)
		if err != nil {
			return
		}

		botMessage, err := FormatAlertHtml(message)
		if err != nil {
			return
		}

		httpLogger.Printf("botMessage: %+v\n", botMessage)
		err = TgBot.Send(botMessage)
		if err != nil {
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	m.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return m
}

type TelegramBot struct {
	TgToken    string
	ChatIds    []string
	bot        *tele.Bot
	recipients []*Recipient
}

func NewTgBot(tgToken string, chatIds ...string) *TelegramBot {
	bot := &TelegramBot{
		TgToken: tgToken,
		ChatIds: chatIds,
	}
	for _, chatId := range chatIds {
		bot.recipients = append(bot.recipients, NewRecipient(chatId))
	}
	return bot
}

func (bot *TelegramBot) Init() (err error) {
	tgBot, err := tele.NewBot(tele.Settings{
		Token:  bot.TgToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	bot.bot = tgBot
	return
}

func (bot *TelegramBot) Send(botMessage string) (err error) {
	for _, r := range bot.recipients {
		_, err = bot.bot.Send(r, botMessage, tele.ModeHTML)
		if err != nil {
			return
		}
	}
	return
}

// ===================================================== alert.go 🔴🟡🟢
var alertHtmlTemplate = `
{{- range $index, $alert := .Alerts}}
=============
告警级别: {{- if eq $alert.Labels.severity "critical" }}🔴{{- else if eq $alert.Labels.severity "warning" }}🟡{{- else}}🟢{{- end}}
告警类型: {{ $alert.Labels.alertname }}
实例信息: {{- if $alert.Labels.job}}{{$alert.Labels.job}}{{- end}} {{- if $alert.Labels.instance}}-{{$alert.Labels.instance}}{{- end}}
{{- if $alert.Annotations.summary }}
告警描述: {{$alert.Annotations.summary}}
{{- end}}
{{- if $alert.Annotations.description }}
告警详情: {{$alert.Annotations.description}}
{{- end}}
故障时间: {{ ($alert.StartsAt.Add 28800e9).Format "2006-01-02 15:04:05" }}
{{- end }}
`

var messageTemplate *template.Template
var tz *time.Location

func init() {
	tz, _ = time.LoadLocation("Asia/Shanghai")
	var err error
	messageTemplate = template.New("").Funcs(template.FuncMap{
		"toUpper": strings.ToUpper,
		"timeFormat": func(t time.Time) string {

			return t.In(tz).Format("Mon, 02 Jan 2006 15:04:05 MST")
		},
		"since": func(t time.Time) string {
			return time.Since(t).Round(time.Second).String()
		},
		"duration": func(start time.Time, end time.Time) string {
			return end.Sub(start).Round(time.Second).String()
		},
	})
	messageTemplate, err = messageTemplate.Parse(alertHtmlTemplate)
	if err != nil {
		panic(err)
	}
}

func FormatAlertHtml(message Message) (string, error) {
	tpl := bytes.Buffer{}

	err := messageTemplate.Execute(&tpl, message)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// ===================================================== recipient.go

func NewRecipient(recipient string) *Recipient {
	return &Recipient{recipient: recipient}
}

type Recipient struct {
	recipient string
}

func (r *Recipient) Recipient() string {
	return r.recipient
}

// ===================================================== alertmanager message
type KV map[string]string
type Alerts []Alert
type Alert struct {
	Status       string    `json:"status"`
	Labels       KV        `json:"labels"`
	Annotations  KV        `json:"annotations"`
	StartsAt     time.Time `json:"startsAt"`
	EndsAt       time.Time `json:"endsAt"`
	GeneratorURL string    `json:"generatorURL"`
	Fingerprint  string    `json:"fingerprint"`
}
type Data struct {
	Receiver string `json:"receiver"`
	Status   string `json:"status"`
	Alerts   Alerts `json:"alerts"`

	GroupLabels       KV `json:"groupLabels"`
	CommonLabels      KV `json:"commonLabels"`
	CommonAnnotations KV `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}
type Message struct {
	*Data

	// The protocol version.
	Version         string `json:"version"`
	GroupKey        string `json:"groupKey"`
	TruncatedAlerts uint64 `json:"truncatedAlerts"`
}
