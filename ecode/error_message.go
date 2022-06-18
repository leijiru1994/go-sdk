package ecode

import "sync"

type errorMessageMap struct {
	sync.RWMutex

	m map[int]string
}

func NewWithMessage(e int, msg string) Code {
	errorMessageM.Lock()
	errorMessageM.m[e] = msg
	errorMessageM.Unlock()

	return Int(e)
}

var errorMessageM errorMessageMap

// 后续这里需要走watch方式，动态更新系统message配置
func init() {
	errorMessageM = errorMessageMap{
		m: map[int]string{
			400: "输入数据有误",
			401: "请在登录后进行该操作",
			429: "前方线路拥堵，请稍后再试",
			500: "服务器异常，请联系客服",
		},
	}
}
