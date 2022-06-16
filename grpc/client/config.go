package client

import (
	"time"
)

type Config struct {
	SrvName string `json:"srv_name" yaml:"srv_name"`
	Target  string `json:"target" yaml:"target"`
	DailTimeout time.Duration `json:"dail_timeout" yaml:"dail_timeout"`
}
