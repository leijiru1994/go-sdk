package tracer

import (
	"encoding/json"
	"log"
	"os"

	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/reporter"
)

type logReporter struct {
	logger *log.Logger
}

func NewReporter(l *log.Logger) reporter.Reporter {
	if l == nil {
		// use standard type of log setup
		l = log.New(os.Stderr, "", log.LstdFlags)
	}
	return &logReporter{
		logger: l,
	}
}

// Send outputs a span to the Go logger.
func (r *logReporter) Send(s model.SpanModel) {
	if bs, err := json.Marshal(s); err == nil {
		r.logger.Printf("%s\n", string(bs))
	}
}

// Close closes the reporter
func (*logReporter) Close() error { return nil }
