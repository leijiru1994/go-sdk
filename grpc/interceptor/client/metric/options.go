package metric

type options struct {
	client string
}

type Option func(*options)

var (
	defaultOptions = &options{}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}

	if optCopy.client == "" {
		optCopy.client = "unknown-grpc-client"
	}

	return optCopy
}

// WithTraceHeaderName customizes the trace header name where trace metadata passed with requests.
func WithClientName(name string) Option {
	return func(o *options) {
		o.client = name
	}
}
