package log

import (
	"context"
	"os"
	"path"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

const (
	maxFileSize = 50 * (2 >> 20) // 50M
)

var (
	serverLogMetric *prometheus.CounterVec
)

func init() {
	serverLogMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mp_server_log",
			Help: "Total number of diode missed log record count.",
		},
		[]string{"server_name", "biz_name"})

	prometheus.MustRegister(serverLogMetric)
}

func InitLogger(ctx context.Context, conf *Config) (logger zerolog.Logger, err error) {
	if conf == nil || conf.File == "" {
		logger = zerolog.New(os.Stderr).With().Logger()
		return
	}

	err = os.MkdirAll(path.Dir(conf.File), 0777)
	if err != nil {
		return
	}

	var f1 *os.File
	f1, err = os.OpenFile(conf.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return
	}

	w1 := diode.NewWriter(f1, 1000000, 100*time.Millisecond, func(missed int) {
		serverLogMetric.WithLabelValues([]string{os.Getenv("server_name"), conf.BizName}...).Inc()
	})


	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	logger = zerolog.New(w1).With().Logger()
	logger.Level(conf.Level)

	//err = startRotateFile(ctx, conf.AccessLogFile)
	//if err != nil {
	//	return
	//}
	//
	//err = startRotateFile(ctx, conf.BizLogFile)

	return
}

// NOTE: 暂时不使用，需要额外再测试一下
func startRotateFile(ctx context.Context, file string) (err error) {
	var rotate *rotatelogs.RotateLogs
	var currentFile string
	rotate, err = rotatelogs.New(
		file,
		rotatelogs.WithHandler(rotatelogs.HandlerFunc(func(e rotatelogs.Event) {
			if e.Type() != rotatelogs.FileRotatedEventType {
				return
			}

			currentFile = e.(*rotatelogs.FileRotatedEvent).CurrentFile()
			go func(previousFile string) {
				// 日志轮转后，两分钟后删除轮转文件，防止轮转过程中logtail日志采集丢失
				time.Sleep(time.Minute * 2)
				_ = os.Remove(previousFile)
			}(e.(*rotatelogs.FileRotatedEvent).PreviousFile())
		})),
	)
	if err != nil {
		return
	}

	currentFile = rotate.CurrentFileName()
	go func() {
		ticker := time.NewTicker(time.Minute * 5)
		defer func() {
			ticker.Stop()
		}()

		for {
			select {
			case <-ticker.C:
				_ = rotateFile(rotate, currentFile)
			case <-ctx.Done():
				return
			}
		}
	}()

	return
}

func rotateFile(rotate *rotatelogs.RotateLogs, file string) (err error) {
	if file == "" {
		return
	}

	info, statErr := os.Stat(file)
	if statErr != nil {
		return
	}

	if info.Size() < maxFileSize {
		return
	}

	err = rotate.Rotate()
	return
}
