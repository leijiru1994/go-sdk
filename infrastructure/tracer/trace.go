package tracer

import (
	innerLog "log"
	"os"
	"time"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	innerReporter "github.com/openzipkin/zipkin-go/reporter"
	"github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/openzipkin/zipkin-go/reporter/log"
)

var (
	GlobalTracer *zipkin.Tracer
)

func InitTracer(conf *Config) (err error) {
	var reporter innerReporter.Reporter
	if conf.ReportUrl == "" {
		if conf.File != "" {
			var f *os.File
			f, err = os.OpenFile(conf.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				return
			}

			reporter = NewReporter(innerLog.New(f, "", 0))
		} else {
			reporter = log.NewReporter(nil)
		}
	} else {
		reporter = http.NewReporter(conf.ReportUrl)
	}

	var sampler zipkin.Sampler
	sampler, err = zipkin.NewBoundarySampler(conf.Rate, time.Now().UnixNano())
	if err != nil {
		return
	}

	var endpoint *model.Endpoint
	endpoint, err = zipkin.NewEndpoint(os.Getenv("PROJECT_NAME"), "")
	if err != nil {
		return
	}

	GlobalTracer, err = zipkin.NewTracer(
		reporter,
		zipkin.WithSampler(sampler),
		zipkin.WithTraceID128Bit(true),
		zipkin.WithLocalEndpoint(endpoint))

	return
}
