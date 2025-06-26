package up7

import (
	"fmt"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

func TestUp7(t *testing.T) {
	// 创建Excel文件
	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetSheetName("Sheet1", sheet)

	// 写表头
	headers := []string{
		"区域ID",
		"描述",
		"局数",
		"赢局数",
		"结束分数",
		"停止条件",
		"所有区域押分",
		"所有区域返还押分",
		"所有区域赢分",
		"所有区域输分",
		"中奖率",
		"RTP",
		"总输赢",
		"总暴击",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)

	}

	// 写数据
	for idx, area := range betAreas {
		res := simulateArea(area)
		row := idx + 2 // Excel行号从2开始
		values := []interface{}{
			res.BetAreaID,
			res.Description,
			res.TotalCount,
			res.TotalWinCount,
			res.FinalMoney,
			res.StopReason,
			res.TotalBet,
			res.TotalRetBet,
			res.TotalWin,
			res.TotalLose,
			res.WinRate,
			res.RTP,
			res.TotalWinLose,
			res.TotalBaoji,
		}
		for col, v := range values {
			cell, _ := excelize.CoordinatesToCellName(col+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	// 保存文件
	if err := f.SaveAs("rtp_result.xlsx"); err != nil {
		fmt.Println("保存Excel失败:", err)
	} else {
		fmt.Println("结果已保存到 rtp_result.xlsx")
	}

}
