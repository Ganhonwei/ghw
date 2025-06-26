package utils

import (
	"fmt"
)

var (
	FormatString = "%v\nthe trace error is\n%s"
)

//按格式返回一个错误
//同时携带原始的错误信息
func NewError(err error, format string, p ...interface{}) error {
	return fmt.Errorf(FormatString, fmt.Sprintf(format, p...), err)
}

//返回一个错误
func New(format string, p ...interface{}) error {
	return fmt.Errorf(format, p...)
}
