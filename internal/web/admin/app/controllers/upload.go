package controllers

import (
	"errors"
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/entity"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
	"github.com/globalsign/mgo/bson"
)

type UploadController struct {
	BaseController
}

type ChildType struct {
	ID   int32
	Name string
}

// 配置上传
func (this *UploadController) Config() {
	// beego.Alert("操作成功")
	// this.Ctx.Output.Script(js, MSG_OK)
	page, _ := this.GetInt("page")
	tabid, _ := this.GetInt("tabid")
	ftype, _ := this.GetInt("ftype")
	remark := this.GetString("remark")
	if page < 1 {
		page = 1
	}
	if tabid == 1 {
		if this.isPost() {
			// 判断是否上传文件
			f, h, err := this.GetFile("filename")
			type_id := this.GetString("type")
			fmt.Println("获取上传文件:", h)
			if remark == "" {
				this.checkError(errors.New("请输入备注!"))
			}
			if err != nil {
				this.checkError(errors.New("请选择正确的上传文件"))
				return
			}
			// 延迟关闭文件
			defer f.Close()

			// 判断文件类型是否为Excel
			fileName := h.Filename
			excelRegex := regexp.MustCompile(`\.(xlsx|xls)$`)
			if !excelRegex.MatchString(fileName) {
				this.checkError(errors.New("请选择Excel文件！"))
				return
			}
			parts := strings.Split(fileName, ".")
			fileNameWithoutExtension := strings.Join(parts[:len(parts)-1], ".")
			fmt.Print(fileNameWithoutExtension)
			// 读取上传文件内容, data就是文件的内容
			// ioutil.ReadAll返回[]byte类型数据
			data, err := ioutil.ReadAll(f)
			if err != nil {
				this.checkError(errors.New("读取上传文件内容错误"))
			}
			msg := new(pb.UploadConfig)
			msg.Id = type_id
			msg.Data = data
			msg.FileName = fileNameWithoutExtension
			result, err := service.GmRequest(pb.WebUploadConfig, pb.ConfigAtype(pb.CONFIG_UPLOAD_CONFIG), msg)
			beego.Trace("result: ", result)
			resmsg := ""
			status := 0
			if err == nil {
				resmsg = "上传成功"
				status = 1
			} else {
				resmsg = err.Error()
			}
			records := new(entity.UploadRecords)
			records.Name = fileNameWithoutExtension
			records.FType = type_id
			records.Status = status
			records.ResMessage = resmsg
			records.Operator = this.auth.GetUser().UserName
			records.Remark = remark
			service.UploadService.AddUploadRecords(records)
			if err != nil {
				this.checkError(err)
			} else {
				service.ActionService.Add("Upload_Config", this.auth.GetUser().UserName, "", type_id, strconv.Itoa(status), remark)
				this.showMsg("上传文件成功！", MSG_OK, "/admin/upload/config")
			}
		}
	} else if tabid == 2 {
		// 更新配置
		if this.isPost() {
			result, err := service.GmRequest(pb.WebReloadConfig, pb.ConfigAtype(pb.CONFIG_UPLOAD_CONFIG), nil)
			beego.Trace("result: ", result)
			if err == nil {
				service.ActionService.Add("Upload_Config_Notify", this.auth.GetUser().UserName, "", "", "", remark)
				this.showMsg("重新加载配置成功！", MSG_OK, "/admin/upload/config")
			}
		}
	} else {
		m := bson.M{}
		if ftype != 0 {
			m["f_type"] = strconv.Itoa(ftype - 1)
		}
		list, _ := service.UploadService.GetUploadRecords(page, this.pageSize, m)
		count, _ := service.UploadService.GetUploadRecordsTotal(m)
		this.Data["count"] = count
		this.Data["list"] = list
		this.Data["pageBar"] = libs.NewPager(page, int(count), this.pageSize, beego.URLFor("UploadController.Config", "tabid", tabid, "ftype", ftype), true).ToString()
		ftypes := map[int]string{
			0:  "全部",
			1:  "GAME",
			2:  "TP",
			3:  "DRAGON TIGER",
			4:  "7UPDOWN",
			5:  "RUMMY",
			6:  "TP-AK47",
			7:  "TP-JOKER",
			8:  "CRASH",
			9:  "ANDARBAHAR",
			10: "彩票",
			11: "飞机",
			12: "RUMMY双人",
		}
		this.Data["ftype"] = ftype
		this.Data["ftypes"] = ftypes
	}
	types := map[int]string{
		0:  "GAME",
		1:  "TP",
		2:  "DRAGON TIGER",
		3:  "7UPDOWN",
		4:  "RUMMY",
		5:  "TP-AK47",
		6:  "TP-JOKER",
		7:  "CRASH",
		8:  "ANDARBAHAR",
		9:  "彩票",
		10: "飞机",
		12: "RUMMY双人",
	}
	this.Data["pageTitle"] = "配置上传"
	this.Data["types"] = types
	this.Data["tabid"] = tabid
	this.display()
}
