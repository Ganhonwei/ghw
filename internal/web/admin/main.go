package admin

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"goserver/internal/web/admin/app/controllers"
	_ "goserver/internal/web/admin/app/mail"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"

	"github.com/astaxie/beego"
	"github.com/beego/i18n"
)

const VERSION = "2.0.1"

func trunc(str string) string {
	strVal := strings.TrimSuffix(str, string([]rune(str)[len(str)-1]))
	strhtml := ""
	if strings.Contains(strVal, ",") {
		substrs := strings.Split(strVal, ",")
		for i, substr := range substrs {
			if strings.Contains(substr, ":") {
				substr1 := strings.Split(substr, ":")
				float, err := strconv.ParseFloat(substr1[1], 64)
				if err == nil {
					money := (float / 100)
					floatStr := strconv.FormatFloat(money, 'f', 2, 64)
					strMsg := substr1[0] + ":" + floatStr
					strhtml += strMsg
				} else {
					if i == 0 {
						strhtml = substr + " | "
					} else {
						strhtml += substr
					}
				}
			} else {
				if i == 0 {
					strhtml = substr + " | "
				} else {
					strhtml += substr
				}
			}

		}
		return strhtml
	} else {
		if strings.Contains(strVal, ":") {
			substr1 := strings.Split(strVal, ":")
			float, err := strconv.ParseFloat(substr1[1], 64)
			if err == nil {
				money := (float / 100)
				floatStr := strconv.FormatFloat(money, 'f', 2, 64)
				strMsg := substr1[0] + ":" + floatStr
				strhtml += strMsg
			} else {
				strhtml += strVal
			}
		} else {
			strhtml += strVal
		}
		return strhtml
	}
}

// 是否是癞子
func IsLaizi(card, laizi uint32) bool {
	if card == laizi {
		return true
	}
	return false
}

func StringCutOff(str string) string {
	strval := ""
	if str != "" {
		if len(str) > 5 {
			strval = str[:5] + "..."
		} else {
			strval = str
		}
	}
	return strval
}

type Service struct{}

func (s *Service) Run() {
	beego.AppConfig.Set("version", VERSION)
	runmode := beego.AppConfig.String("runmode")
	service.RunMode = runmode

	// service.InitNsq()
	service.InitNats()

	//service.Init()
	service.InitMgo()

	// nats 请求路由
	service.InitNatsRouter()
	utils.InitSnowflakeId(0)

	service.InitClickhouse()

	if runmode == "dev" {
		beego.SetLevel(beego.LevelDebug)
	} else {
		beego.SetLevel(beego.LevelInformational)
		beego.SetLogger("file", `{"filename":"`+beego.AppConfig.String("log_file")+`"}`)
		beego.BeeLogger.DelLogger("console")
	}

	// 记录启动时间
	beego.AppConfig.Set("up_time", fmt.Sprintf("%d", time.Now().Unix()))
	beego.AddFuncMap("trunc", trunc)
	beego.AddFuncMap("IsLaizi", IsLaizi)
	beego.AddFuncMap("StringCutOff", StringCutOff)
	beego.AddFuncMap("i18n", i18n.Tr)

	/*
		beego.Router("/", &controllers.MainController{}, "*:Index")
		beego.Router("/login", &controllers.MainController{}, "*:Login")
		beego.Router("/logout", &controllers.MainController{}, "*:Logout")
		beego.Router("/profile", &controllers.MainController{}, "*:Profile")
		beego.Router("/regist", &controllers.MainController{}, "*:Regist")
		beego.Router("/servers", &controllers.MainController{}, "*:Servers")
		beego.Router("/files", &controllers.MainController{}, "*:Files")

		beego.AutoRouter(&controllers.UserController{})
		beego.AutoRouter(&controllers.RoleController{})
		beego.AutoRouter(&controllers.MailTplController{})
		beego.AutoRouter(&controllers.MainController{})
		beego.AutoRouter(&controllers.PlayerController{})
		beego.AutoRouter(&controllers.LoggerController{})
		beego.AutoRouter(&controllers.AgencyController{})
		beego.AutoRouter(&controllers.ChartsController{})

		beego.ErrorController(&controllers.ErrorController{})

		beego.SetStaticPath("/assets", "assets")
		beego.SetStaticPath("/contract", "views/main/contract.html")
		beego.SetStaticPath("/rules", "views/main/rules.html")
		beego.SetStaticPath("/download", "views/main/download.html")
		//beego.SetStaticPath("/headimag", "headimag")
		beego.SetStaticPath("/poster", "views/main/poster.html")
	*/

	//namespace
	namespace := beego.AppConfig.String("namespace")
	beego.Trace("namespace: ", namespace)

	//初始化 namespace
	ns :=
		beego.NewNamespace("/"+namespace,

			beego.NSRouter("/", &controllers.MainController{}, "*:Index"),
			beego.NSRouter("/login", &controllers.MainController{}, "*:Login"),
			beego.NSRouter("/logout", &controllers.MainController{}, "*:Logout"),
			beego.NSRouter("/profile", &controllers.MainController{}, "*:Profile"),
			beego.NSRouter("/regist", &controllers.MainController{}, "*:Regist"),
			beego.NSRouter("/servers", &controllers.MainController{}, "*:Servers"),
			beego.NSRouter("/files", &controllers.MainController{}, "*:Files"),

			beego.NSRouter("/player", &controllers.PlayerController{}, "*:List"),

			beego.NSAutoRouter(&controllers.ActionController{}),
			beego.NSAutoRouter(&controllers.UserController{}),
			beego.NSAutoRouter(&controllers.RoleController{}),
			beego.NSAutoRouter(&controllers.MailTplController{}),
			beego.NSAutoRouter(&controllers.MainController{}),
			beego.NSAutoRouter(&controllers.PlayerController{}),
			beego.NSAutoRouter(&controllers.LoggerController{}),
			beego.NSAutoRouter(&controllers.AgencyController{}),
			beego.NSAutoRouter(&controllers.ChartsController{}),

			beego.NSAutoRouter(&controllers.PayController{}),
			beego.NSAutoRouter(&controllers.GameController{}),
			beego.NSAutoRouter(&controllers.MessageController{}),
			beego.NSAutoRouter(&controllers.Statistics2Controller{}),
			beego.NSAutoRouter(&controllers.StatisticsController{}),
			beego.NSAutoRouter(&controllers.GameStatsController{}),
			beego.NSAutoRouter(&controllers.UploadController{}),
			beego.NSAutoRouter(&controllers.ActivityController{}),
			beego.NSAutoRouter(&controllers.ChannelController{}),
		)
	//注册 namespace
	beego.AddNamespace(ns)

	beego.ErrorController(&controllers.ErrorController{})

	beego.SetStaticPath("/assets", "assets")
	// beego.SetStaticPath("/"+namespace+"/assets", "assets")
	beego.SetStaticPath("/"+namespace+"/contract", "views/main/contract.html")
	beego.SetStaticPath("/"+namespace+"/termsofservice", "views/main/termsofservice.html")
	beego.SetStaticPath("/"+namespace+"/privacypolicy", "views/main/privacypolicy.html")
	beego.SetStaticPath("/"+namespace+"/agencyAgreement", "views/main/agencyAgreement.html")
	//beego.SetStaticPath("/admin/rules", "views/main/rules.html")
	//beego.SetStaticPath("/admin/download", "views/main/download.html")
	//beego.SetStaticPath("/admin/headimag", "headimag")
	beego.SetStaticPath("/"+namespace+"/poster", "views/main/poster.html")

	beego.Run()
}
