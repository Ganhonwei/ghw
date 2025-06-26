package handler

import (
	"goserver/pkg/utils"
	"strconv"
	"strings"
)

type stringslice struct {
	Value []string
}

func (s *stringslice) UnmarshalXLS(input string) error {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		s.Value = append(s.Value, part)
	}
	return nil
}

type ints struct {
	Value []int
}

func (i *ints) UnmarshalXLS(input string) error {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		if part == "" {
			continue
		}
		// 将字符串转换为整数
		num, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		i.Value = append(i.Value, num)
	}
	return nil
}

type intss struct {
	Value [][]int
}

func (iss *intss) UnmarshalXLS(input string) error {
	parts1 := strings.Split(input, ":")
	for _, part1 := range parts1 {
		part1 = strings.Trim(part1, "[]")
		parts2 := strings.Split(part1, ",")
		var v []int
		for _, part2 := range parts2 {
			num, err := strconv.Atoi(part2)
			if err != nil {
				return err
			}
			v = append(v, num)
		}
		iss.Value = append(iss.Value, v)
	}
	return nil
}

type int32s struct {
	Value []int32
}

func (i *int32s) UnmarshalXLS(input string) error {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		// 将字符串转换为整数
		num, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		i.Value = append(i.Value, int32(num))
	}
	return nil
}

type int64s struct {
	Value []int64
}

func (i *int64s) UnmarshalXLS(input string) error {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		// 将字符串转换为整数
		num, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		i.Value = append(i.Value, int64(num))
	}
	return nil
}

type int64ss struct {
	Value [][]int64
}

func (iss *int64ss) UnmarshalXLS(input string) error {
	parts1 := strings.Split(input, ":")
	for _, part1 := range parts1 {
		part1 = strings.Trim(part1, "[]")
		parts2 := strings.Split(part1, ",")
		var v []int64
		for _, part2 := range parts2 {
			num, err := strconv.Atoi(part2)
			if err != nil {
				return err
			}
			v = append(v, int64(num))
		}
		iss.Value = append(iss.Value, v)
	}
	return nil
}

type int32ss struct {
	Value [][]int32
}

func (iss *int32ss) UnmarshalXLS(input string) error {
	parts1 := strings.Split(input, ":")
	for _, part1 := range parts1 {
		part1 = strings.Trim(part1, "[]")
		parts2 := strings.Split(part1, ",")
		var v []int32
		for _, part2 := range parts2 {
			num, err := strconv.Atoi(part2)
			if err != nil {
				return err
			}
			v = append(v, int32(num))
		}
		iss.Value = append(iss.Value, v)
	}
	return nil
}

type uint32s struct {
	Value []uint32
}

func (i *uint32s) UnmarshalXLS(input string) error {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		// 将字符串转换为整数
		num, err := strconv.Atoi(part)
		if err != nil {
			return err
		}
		i.Value = append(i.Value, uint32(num))
	}
	return nil
}

type float64s struct {
	Value []float64
}

func (i *float64s) UnmarshalXLS(input string) error {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	for _, part := range parts {
		i.Value = append(i.Value, utils.Float64(part))
	}
	return nil
}
