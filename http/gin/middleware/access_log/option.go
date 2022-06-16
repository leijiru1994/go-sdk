package access_log

import (
	"github.com/rs/zerolog"
)

type options struct {
	init bool
	logger zerolog.Logger
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

	if !optCopy.init {
		return nil
	}

	return optCopy
}

// WithLogger customizes the logger.
func WithLogger(logger zerolog.Logger) Option {
	return func(o *options) {
		o.logger = logger
		o.init = true
	}
}
