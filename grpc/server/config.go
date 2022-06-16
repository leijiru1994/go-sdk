package server

import (
	"fmt"
	"net/http"

	"github.com/facebookgo/grace/gracenet"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

type Config struct {
	AccessLog string        `yaml:"access_log"`
	LogLevel  zerolog.Level `yaml:"log_level"` // debug: 0 info: 1 warn: 2 error: 3 fatal: 4 panic: 5

	HttpPort  string `yaml:"http_port"`
	GrpcPort  string `yaml:"http_port"`
	PidFile   string `yaml:"pid_file"`
	CanKill   bool   `yaml:"can_kill"` // 如果用systemd托管，这里请设置为true
}

type Server struct {
	HttpAddr string
	GrpcAddr string
	PidFile  string
	S1       *grpc.Server
	S2       *http.Server
	net      *gracenet.Net
	SMux     *runtime.ServeMux
	CanKill  bool   // 如果用systemd托管，这里请设置为true
}

func (s *Server) SetHttpServer(server *http.Server) {
	s.S2 = server
}

func (s *Config) GetHttpAddr() string {
	if s.HttpPort == "" {
		s.HttpPort = "9000"
	}

	return fmt.Sprintf(":%v", s.HttpPort)
}

func (s *Config) GetGRPCAddr() string {
	if s.GrpcPort == "" {
		s.GrpcPort = "9001"
	}

	return fmt.Sprintf(":%v", s.GrpcPort)
}
