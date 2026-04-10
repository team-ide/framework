package framework

import (
	"encoding/json"
)

type TError interface {
	error
	GetCode() string
	GetMsg() string
}

type CoreError struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
}

func (this_ *CoreError) GetCode() string {
	return this_.Code
}

func (this_ *CoreError) GetMsg() string {
	return this_.Msg
}

func (this_ *CoreError) Error() string {
	bs, _ := json.Marshal(this_)
	return string(bs)
}

// NewError 构造异常对象，code为错误码，msg为错误信息
func NewError(code string, msg string) *CoreError {
	err := &CoreError{
		Code: code,
		Msg:  msg,
	}
	return err
}
