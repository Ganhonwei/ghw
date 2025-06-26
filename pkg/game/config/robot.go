package config

import (
	"fmt"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

// 名字
func GetRobotNameByExcel(sheet string, f *excelize.File) ([]string, []string, error) {
	manRows, err1 := f.GetRows(sheet)
	if err1 != nil {
		fmt.Println(err1)
		return nil, nil, err1
	}
	manEnNames := make([]string, 0) // 英文名
	manInNames := make([]string, 0) // 印度名
	for _, v := range manRows {
		if v[0] != "" {
			manEnNames = append(manEnNames, v[0])
		}
		if len(v) > 1 && v[1] != "" {
			manInNames = append(manInNames, v[1])
		}
	}
	return manEnNames, manInNames, nil
}
