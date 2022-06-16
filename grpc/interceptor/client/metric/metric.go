package metric

import (
	"context"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

import "github.com/prometheus/client_golang/prometheus"

var GPRCClientLatencyMetric *prometheus.HistogramVec

func init() {
	GPRCClientLatencyMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "grpc_client_handling_seconds",
			Help: "Histogram of response latency (seconds) of grpc that had been application-level handled by the client.",
			Buckets: []float64{.005, .01, .025, .05, .075, .1, .25, .5, .75, 1},
		},
		[]string{"grpc_client", "grpc_type", "grpc_service", "grpc_method", "grpc_code"})
	prometheus.MustRegister(GPRCClientLatencyMetric)
}

// UnaryServerInterceptor returns a new unary client interceptor for OpenTracing.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	opt := evaluateOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startedAt := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		st, _ := status.FromError(err)
		serviceName, methodName := splitMethodName(method)
		GPRCClientLatencyMetric.WithLabelValues(
			opt.client,
			string(grpc_prometheus.Unary),
			serviceName,
			methodName,
			st.Code().String(),
		).Observe(time.Since(startedAt).Seconds())

		return err
	}
}

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}
