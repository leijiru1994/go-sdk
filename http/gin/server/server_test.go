package server

import (
	"context"
	"github.com/rs/zerolog"
	"github.com/leijjiru1994/go-sdk/http/gin/middleware/access_log"
	"github.com/leijjiru1994/go-sdk/http/gin/middleware/metric"
	"github.com/leijjiru1994/go-sdk/infrastructure/log"
	"testing"
	"time"

	"github.com/leijjiru1994/go-sdk/http/gin/common"
	"github.com/leijjiru1994/go-sdk/infrastructure/tracer"

	"github.com/gin-gonic/gin"
	"github.com/rs/cors"
)

func TestHttpServer(t *testing.T) {
	var logger zerolog.Logger
	logger, err := log.InitLogger(context.Background(), &log.Config{
		BizName: "access_log",
		File:    "",
		Level:   zerolog.DebugLevel,
	})
	if err != nil {
		t.Error(err)
		return
	}

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
			//Debug: true,
		},
		HandlerFunc: []gin.HandlerFunc{
			metric.Metrics(),
			access_log.WithAccessLog(access_log.WithLogger(logger)),
			gin.Recovery(),
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
