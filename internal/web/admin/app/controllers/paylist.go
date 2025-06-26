package controllers

import (
	"fmt"
	"goserver/gen/pb"
	"goserver/internal/web/admin/app/args"
	"goserver/internal/web/admin/app/libs"
	"goserver/internal/web/admin/app/service"
	"goserver/pkg/utils"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/astaxie/beego"
)

// 充值列表
func (c *PayController) PayList() {
	action := c.GetString("action")

	if action == "Aside" {
		// 支付渠道
		payChannels, err := service.GameService.GetPayChannel()
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		var channels = []map[string]any{}
		for _, c := range payChannels {
			id, err := strconv.Atoi(c.Id)
			if err == nil {
				channels = append(channels, map[string]any{
					"id":   id,
					"name": c.Name,
				})
			}
		}

		// 渠道、渠道类
		packages, _ := service.ChannelService.GetChannelAll(false)
		packageClasses, _ := service.ChannelService.GetChannelClassAll(false)

		c.JsonRSuccess(libs.R{
			"channels":       channels,
			"packages":       packages,
			"packageClasses": packageClasses,
			"isOperation":    c.auth.HasAccessPerm(c.controllerName, "paylistopt"),
			"isExport":       c.auth.HasAccessPerm(c.controllerName, "paylistexport"),
			"version":        service.GetVersions(),
			"today":          service.NowTime().Format(utils.FORMAT_DATE),
		})
		return
	}

	if action == "Stats" {
		args := new(args.PayListArgs)
		c.JsonBody(args)
		ret, err := service.PayService.PayListStats(*args)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}
		c.JsonRSuccess(ret)
		return
	}

	if action == "PayList" {
		args := new(args.PayListArgs)
		c.JsonBody(args)
		page, pageSize, sortBy, asc := c.PageArgs()
		if sortBy != "" {
			if !utils.SliceIn(sortBy, "ctime", "firstPay", "amount") {
				c.JsonRFail("排序字段有误")
				return
			}
		} else {
			sortBy = "ctime"
			asc = 0
		}
		total, list, err := service.PayService.PayList(page, pageSize, sortBy, asc, *args, true)
		if err != nil {
			c.JsonRError(err.Error())
			return
		}

		c.JsonRSuccess(libs.R{
			"total": total,
			"list":  list,
		})
		return
	}

	c.Data["pageTitle"] = "充值记录"
	c.display()
}

// 充值记录(操作)
func (c *PayController) PayListOpt() {
	action := c.GetString("action")
	orderid := c.GetString("id")
	if orderid == "" {
		c.JsonRFail("订单ID不能为空")
		return
	}
	userName := c.auth.GetUser().UserName

	// 手动补单
	if action == "PayCallback" {
		service.PayService.UpdatePayRepairAdmin(orderid, userName)

		result, err := service.GmRequest(pb.WebPayCallback, pb.CONFIG_UPSERT, orderid)
		beego.Trace("PayCallback result: ", result)
		service.ActionService.Add("pay_callback", userName,
			"", utils.String(orderid), utils.String(orderid), "")

		if err != nil {
			c.JsonRError(err.Error())
			return
		} else {
			c.JsonRSuccess(nil)
			return
		}
	}

	record := service.PayService.GetPayRecord(orderid)
	if record.OrderID == "" {
		c.JsonRFail("订单不存在")
		return
	}

	// 添加标记
	if action == "PayTag" {
		tag := c.GetString("tag")
		var errorMsgs []string
		if record.ErrorMsg != "" {
			errorMsgs = strings.Split(record.ErrorMsg, ";")
		}
		if len(errorMsgs) >= 5 {
			c.JsonRFail("订单标记信息超过5条")
			return
		}
		tag = strings.ReplaceAll(tag, ";", "；")

		// 2025/1/15 15：22：33 admin：通道查实未收到
		errorMsg := fmt.Sprintf("%s %s:%s", service.NowTime().Format(utils.FORMAT2), userName, tag)
		errorMsgs = append(errorMsgs, errorMsg)
		ok := service.PayService.UpdatePayTag(orderid, strings.Join(errorMsgs, ";"))
		if !ok {
			c.JsonRError("标记失败")
			return
		} else {
			c.JsonRSuccess(nil)
			return
		}
	}

	// 删除标记
	if action == "PayTagDel" {
		tagIndex, err := c.GetInt("tagIndex")
		if err != nil {
			c.JsonRFail("标记索引有误")
			return
		}
		var errorMsgs []string
		if record.ErrorMsg != "" {
			errorMsgs = strings.Split(record.ErrorMsg, ";")
		}
		utils.SliceReverse(errorMsgs)

		var newErrorMsgs []string
		for i, errorMsg := range errorMsgs {
			if i != tagIndex {
				newErrorMsgs = append(newErrorMsgs, errorMsg)
			} else {
				beego.Info("PayTagDel", orderid, errorMsg)
			}
		}
		utils.SliceReverse(newErrorMsgs)

		ok := service.PayService.UpdatePayTag(orderid, strings.Join(newErrorMsgs, ";"))
		if !ok {
			c.JsonRError("删除标记失败")
			return
		} else {
			c.JsonRSuccess(nil)
			return
		}
	}

	c.JsonRFail("unknown request")
}

// 充值记录导出
func (c *PayController) PayListExport() {
	args := new(args.PayListArgs)
	c.JsonBody(args)
	_, _, sortBy, asc := c.PageArgs()
	if sortBy != "" {
		if !utils.SliceIn(sortBy, "ctime", "firstPay", "amount") {
			c.Ctx.Output.SetStatus(http.StatusBadRequest)
			c.JsonRFail("排序字段有误")
			return
		}
	} else {
		sortBy = "ctime"
		asc = 0
	}
	_, list, err := service.PayService.PayList(1, 100000, sortBy, asc, *args, false)
	if err != nil {
		beego.Error(err)
		c.Ctx.Output.SetStatus(http.StatusInternalServerError)
		c.JsonRError(err.Error())
		return
	}

	var rows [][]any
	headers := []string{"UID", "昵称", "渠道别名", "渠道类", "支付通道", "通道类型", "订单号", "第三方订单号", "订单生成时间", "完成时间", "商品名称", "商品类型", "是否首充", "充值金额", "实际支付金额", "到账cash", "赠送cash金额", "赠送bonus金额", "订单状态", "UTR", "标记", "手动补单信息"}
	for _, v := range list {
		row := []any{
			v.Userid,
			v.NickName,
			v.FPackageAlias,
			v.FPackageClass,
			v.FChannel,
			v.FChannelType,
			v.OrderID,
			v.OutTradeNo,
			v.FCtime,
			v.FPayTime,
			v.ShopName,
			v.FShopType,
			v.FFirstPay,
			v.FAmount,
			v.FRealAmount,
			v.FCash,
			v.FGiveCash,
			v.FOtherPresent,
			v.FOrderStatus,
			strings.Join(v.FUtrs, "\n"),
			strings.Join(v.ErrorMsgs, "\n"),
			v.FRepair,
		}
		rows = append(rows, row)
	}

	// 检查是否有数据可导出
	if len(rows) == 0 {
		c.Ctx.Output.SetStatus(http.StatusRequestedRangeNotSatisfiable) // 416请求范围无效
		c.JsonRFail("未查询到可以导出的数据")
		return
	}

	file := excelize.NewFile()
	sheet := "Sheet1"

	//设置表格头
	for i, head := range headers {
		no := NumberToExcelColumn(i + 1)
		file.SetCellValue(sheet, fmt.Sprintf("%s1", no), head)
		if head == "" {
			file.MergeCell("Sheet1", fmt.Sprintf("%s1", no), fmt.Sprintf("%s%d", no, len(rows)+1))
		}
	}
	// 写入数据
	for rno, row := range rows {
		for cno, value := range row {
			column := NumberToExcelColumn(cno + 1)
			// Excel 行号从 1 开始，因此要加 2
			file.SetCellValue(sheet, fmt.Sprintf("%s%d", column, rno+2), value)
		}
	}

	//构造文件名称
	title := "充值记录"
	fileName := title + "_" + time.Now().Format("20060102150405") + ".xlsx"
	// 解决文件名中文乱码问题
	fileName = path.Base(fileName)
	fileName = url.QueryEscape(fileName)
	c.Ctx.Output.Header("Content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Ctx.Output.Header("Content-Disposition", "attachment;filename="+fileName)
	c.Ctx.Output.Header("Pragma", "No-cache")
	c.Ctx.Output.Header("Cache-Control", "No-cache")
	c.Ctx.Output.Header("Expires", "0")
	// var buffer bytes.Buffer
	if err := file.Write(c.Ctx.ResponseWriter); err != nil {
		beego.Error(err)
	}
}
