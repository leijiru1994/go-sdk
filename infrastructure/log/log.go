package log

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/leijjiru1994/go-sdk/common/caller"
)

var (
	GlobalBizLog zerolog.Logger
)

func InitBizLog(ctx context.Context, conf *Config) (err error) {
	if conf == nil {
		return
	}

	GlobalBizLog, err = InitLogger(ctx, conf)
	if err != nil {
		return
	}

	return
}

func ResetLevel(level zerolog.Level) {
	GlobalBizLog.Level(level)
}

func Debug(ctx context.Context, label string, in interface{}) {
	logBizInfo(ctx, zerolog.DebugLevel, label, in)
}

func Info(ctx context.Context, label string, in interface{}) {
	logBizInfo(ctx, zerolog.InfoLevel, label, in)
}

func Warn(ctx context.Context, label string, in interface{}) {
	logBizInfo(ctx, zerolog.WarnLevel, label, in)
}

func Error(ctx context.Context, label string, in interface{}) {
	logBizInfo(ctx, zerolog.ErrorLevel, label, in)
}

func logBizInfo(ctx context.Context, level zerolog.Level, label string, in interface{}) {
	var event *zerolog.Event
	switch level {
	case zerolog.DebugLevel:
		event = GlobalBizLog.Debug()
	case zerolog.InfoLevel:
		event = GlobalBizLog.Info()
	case zerolog.WarnLevel:
		event = GlobalBizLog.Warn()
	case zerolog.ErrorLevel:
		event = GlobalBizLog.Error().Caller(2)
	}

	if event != nil {
		event.Int64("time", time.Now().Unix()).
			Str("trace_id", caller.GetTraceIDFromContext(ctx)).
			Interface("message", in).
			Str("label", label).Send()
	}
}
