package handler

import (
	"bytes"
	"fmt"
	"goserver/pkg/glog"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	decxls "github.com/dhushon/decxls"
)

type CustomFloatUnmarshal struct {
	Value   string
	Percent []int
}

func (c *CustomFloatUnmarshal) UnmarshalXLS(value string) error {
	c.Value = value
	// f , err := toFloat(value)
	// if err != nil {
	// 	return err
	// }
	// c.Percent = f
	// return nil
	input := value
	input = strings.TrimPrefix(input, "[")
	input = strings.TrimSuffix(input, "]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		// 将字符串转换为整数
		num, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		c.Percent = append(c.Percent, num)
	}
	return nil
}

// 测试
func TestGame(t *testing.T) {

	file, err := os.Open("1.TP.xlsx")
	if err != nil {
		t.Log(err)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Log(err)
		return
	}
	byte_reader := bytes.NewReader(data)

	f, err := excelize.OpenReader(byte_reader)
	if err != nil {
		t.Log(err)
		return
	}

	tp1 := []TP1{}

	err = decxls.UnmarshalExcelize(f, "1.房间基础配置表", &tp1)
	if err != nil {
		t.Log(err)
		return
	}

	tp2 := []TP2{}
	err = decxls.UnmarshalExcelize(f, "2.房间发牌配置表", &tp2)
	if err != nil {
		t.Log(err)
		return
	}

	tp3 := []TP3{}
	err = decxls.UnmarshalExcelize(f, "3.发牌牌型配置", &tp3)
	if err != nil {
		t.Log(err)
		return
	}

	tp4 := []TP4{}
	err = decxls.UnmarshalExcelize(f, "4.人机策略组配置", &tp4)
	if err != nil {
		t.Log(err)
		return
	}

	tp5 := []TP5{}
	err = decxls.UnmarshalExcelize(f, "5.人机策略配置", &tp5)
	if err != nil {
		t.Log(err)
		return
	}

	// fmt.Printf("Filename based, sheet selection %s test: %v\n", "1.1", g)

	// g[0].Scountdown.Percent
	// var games []Game
	// if err := csvutil.Unmarshal(data, &games); err != nil {
	// 	fmt.Println("error:", err)
	// }

	// for _, u := range games {
	// 	fmt.Printf("%+v\n", u)
	// }

}

func Test(t *testing.T) {
	m := math.Pow(0.993, 50000/10000)
	diamond := int64(math.Max(5400, 0))
	ret, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(diamond)/50000/m), 64)

	glog.Info("ret:", ret)
}
