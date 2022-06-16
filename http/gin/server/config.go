package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
	"go-sdk/infrastructure/tracer"
	"net/http"
	"time"

	"github.com/facebookgo/grace/gracenet"
	"github.com/rs/zerolog"
)

type Config struct {
	Tracer    *tracer.Config `yaml:"tracer"`
	AccessLog string         `yaml:"access_log"`
	LogLevel  zerolog.Level  `yaml:"log_level"` // debug: 0 info: 1 warn: 2 error: 3 fatal: 4 panic: 5

	HttpPort  string `yaml:"http_port"`
	PidFile   string `yaml:"pid_file"`
	CanKill   bool   `yaml:"can_kill"` // 如果用systemd托管，这里请设置为true

	ReadTimeout  time.Duration `yaml:"Read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	Cors cors.Options `yaml:"cors"` // 跨域配置
	HandlerFunc []gin.HandlerFunc
}

type Server struct {
	HttpAddr string
	PidFile  string
	Engine   *gin.Engine
	S        *http.Server
	net      *gracenet.Net
	CanKill  bool   // 如果用systemd托管，这里请设置为true

	ReadTimeout  time.Duration `yaml:"Read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	Cors cors.Options `yaml:"cors"` // 跨域配置
}

func (s *Config) GetHttpAddr() string {
	if s.HttpPort == "" {
		s.HttpPort = "9000"
	}

	return fmt.Sprintf(":%v", s.HttpPort)
}
