package framework

import (
	"fmt"
)

type TError interface {
	error
	GetCode() string
	GetMsg() string
}

type CoreError struct {
	code string
	msg  string
}

func (this_ *CoreError) GetCode() string {
	return this_.code
}

func (this_ *CoreError) GetMsg() string {
	return this_.msg
}

func (this_ *CoreError) Error() string {
	return fmt.Sprintf("code:%s , msg:%s", this_.code, this_.msg)
}

// NewError 构造异常对象，code为错误码，msg为错误信息
func NewError(code string, msg string) *CoreError {
	err := &CoreError{
		code: code,
		msg:  msg,
	}
	return err
}
