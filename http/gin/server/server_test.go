package server

import (
	"context"
	"testing"
	"time"

	"go-sdk/http/gin/common"
	"go-sdk/infrastructure/tracer"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func TestHttpServer(t *testing.T) {
	cfg := &Config{
		Tracer:       &tracer.Config{
			File:      "",
			ReportUrl: "",
			Rate:      1,
		},
		AccessLog:    "",
		LogLevel:     0,
		HttpPort:     "9000",
		PidFile:      "",
		CanKill:      false,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second*10,
		Cors:         cors.Options{
			AllowedOrigins: []string{"http://www.baidu.com"},
		},
	}

	server, err := NewDefaultServer(context.Background(), cfg)
	if err != nil {
		t.Error(err)
		return
	}

	server.Engine.Any("hello", func(ctx *gin.Context) {
		common.SendOut(ctx, "OK")
	})
	server.Engine.Any("delay", func(ctx *gin.Context) {
		time.Sleep(time.Second*20)
		common.SendOut(ctx, "delay")
	})

	err = server.Run(context.Background())
	t.Error(err)
}
