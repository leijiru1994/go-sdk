package time_control

import (
	"time"
)

type options struct {
	timeout time.Duration
}

type Option func(*options)

var (
	defaultOptions = &options{
		timeout: nil,
	}
)

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}

	if optCopy.timeout == 0 {
		optCopy.timeout = time.Minute
	}

	return optCopy
}

// WithTimeout customizes the timeout for every requests.
func WithTimeout(timeout time.Duration) Option {
	return func(o *options) {
		if timeout == 0 {
			timeout = time.Second*10
		}
		o.timeout = timeout
	}
}
