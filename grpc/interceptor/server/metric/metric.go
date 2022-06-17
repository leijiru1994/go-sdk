package metric

import (
	"context"
	"strconv"
	"time"

	"github.com/leijiru1994/go-sdk/common/caller"
	"github.com/leijiru1994/go-sdk/common/model"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	GrpcServerHandlingHistogramMetric *prometheus.HistogramVec
)

func init() {
	GrpcServerHandlingHistogramMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_handling_seconds",
			Help:    "Histogram of response latency (seconds) of gRPC that had been application-level handled by the server.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"grpc_caller", "grpc_biz_code", "grpc_service", "grpc_method"})
	prometheus.MustRegister(GrpcServerHandlingHistogramMetric)
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var BizCode uint64
		startTime := time.Now()
		resp, err := handler(ctx, req)

		if r, ok := resp.(model.CommonResponse); ok {
			BizCode = uint64(r.GetCode())
			service, method := caller.SplitMethodName(info.FullMethod)
			duration := time.Since(startTime).Seconds()
			GrpcServerHandlingHistogramMetric.WithLabelValues(
				[]string{
					caller.GetMPService(ctx),
					strconv.FormatUint(BizCode, 10),
					service,
					method}...).
				Observe(duration)
		}

		return resp, err
	}
}
