package ecode

import (
	"strconv"

	"github.com/pkg/errors"
)

func New(e int) Code {
	return Int(e)
}

func add(e int) Code {
	return Int(e)
}

type Codes interface {
	Error() string
	Code() int
	Message() string
	Equal(error) bool
}

type Code int

func (e Code) Error() string {
	return strconv.FormatInt(int64(e), 10)
}

func (e Code) Code() int {
	return int(e)
}

func (e Code) Message() string {
	return e.Error()
}

func (e Code) Equal(err error) bool {
	return EqualError(e, err)
}

func Int(i int) Code {
	return Code(i)
}

func String(e string) Code {
	if e == "" {
		return OK
	}

	i, err := strconv.Atoi(e)
	if err != nil {
		return ServerErr
	}

	return Code(i)
}

func Cause(e error) Codes {
	if e == nil {
		return OK
	}

	if d, ok := errors.Cause(e).(Codes); ok {
		return d
	}

	return String(e.Error())
}

func Equal(a, b Codes) bool {
	if a == nil {
		a = OK
	}

	if b == nil {
		b = OK
	}

	return a.Code() == b.Code()
}

func EqualError(code Codes, err error) bool {
	return Cause(err).Code() == code.Code()
}
