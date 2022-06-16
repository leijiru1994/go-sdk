package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leijjiru1994/go-sdk/grpc/interceptor/client/metric"
	"github.com/leijjiru1994/go-sdk/grpc/interceptor/client/trace"
	"github.com/leijjiru1994/go-sdk/infrastructure/tracer"

	"google.golang.org/grpc"
)

// NOTE: 默认支持metric与trace
// tracer先初始化，否则trace拦截器无效
func NewClientConn(conf Config, opts ...grpc.DialOption) (cc *grpc.ClientConn, err error) {
	if conf.SrvName == "" || conf.Target == "" {
		err = errors.New("srvName or target can not empty")
		return
	}

	if conf.DailTimeout == 0 {
		conf.DailTimeout = 2*time.Second
	}

	ctx4GRPC, _ := context.WithTimeout(context.Background(), conf.DailTimeout)
	if opts == nil || len(opts) == 0 {
		opts = []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithInsecure(),
			grpc.WithChainUnaryInterceptor(
				metric.UnaryServerInterceptor(
					metric.WithClientName(conf.SrvName),
				),
				trace.UnaryServerInterceptor(
					trace.WithClientName(conf.SrvName),
					trace.WithTracer(tracer.GlobalTracer),
				),
			),
		}
	}

	cc, err = grpc.DialContext(ctx4GRPC, conf.Target, opts...)
	if err != nil {
		err = errors.New(fmt.Sprintf("grpc conn err: %v, host: %v", err, conf.Target))
	}
	return
}
