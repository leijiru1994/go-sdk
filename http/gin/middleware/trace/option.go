package trace

import "github.com/openzipkin/zipkin-go"

type options struct {
	client string
	tracer *zipkin.Tracer
}

type Option func(*options)

var (
	defaultOptions = &options{
		tracer: nil,
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}

	if optCopy.client == "" {
		optCopy.client = "unknown-client"
	}

	return optCopy
}

// WithTraceHeaderName customizes the trace header name where trace metadata passed with requests.
func WithClientName(name string) Option {
	return func(o *options) {
		o.client = name
	}
}

// WithTracer sets a custom tracer to be used for this middleware.
func WithTracer(tracer *zipkin.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}
