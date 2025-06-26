package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// 印度主要ISP的IP地址范围
var indiaIPRanges = []struct {
	start string
	end   string
}{
	{"115.240.0.0", "115.255.255.255"}, // Bharti Airtel
	{"116.202.0.0", "116.202.255.255"}, // Tata Communications
	{"122.160.0.0", "122.175.255.255"}, // Bharti Airtel
	{"125.16.0.0", "125.31.255.255"},   // Reliance Communications
	{"180.87.0.0", "180.87.255.255"},   // Tata Communications
	{"182.66.0.0", "182.79.255.255"},   // Bharti Airtel
}

// GenerateIndianIP 生成一个随机的印度IP地址
func GenerateIndianIP() string {
	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 随机选择一个IP范围
	ipRange := indiaIPRanges[rand.Intn(len(indiaIPRanges))]

	// 将IP地址转换为整数
	start := ipToInt(ipRange.start)
	end := ipToInt(ipRange.end)

	// 生成随机IP
	randomIP := start + rand.Int63n(end-start+1)

	// 将整数转换回IP地址
	return intToIP(randomIP)
}

// ipToInt 将IP地址转换为整数
func ipToInt(ip string) int64 {
	var a, b, c, d int64
	_, err := fmt.Sscanf(ip, "%d.%d.%d.%d", &a, &b, &c, &d)
	if err != nil {
		return 0
	}
	return (a << 24) | (b << 16) | (c << 8) | d
}

// intToIP 将整数转换为IP地址
func intToIP(ipInt int64) string {
	a := (ipInt >> 24) & 0xFF
	b := (ipInt >> 16) & 0xFF
	c := (ipInt >> 8) & 0xFF
	d := ipInt & 0xFF
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}
