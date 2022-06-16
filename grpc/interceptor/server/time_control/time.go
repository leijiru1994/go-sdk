package time_control

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	opt := evaluateOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// if deadline is nil , injection it
		if _, ok := ctx.Deadline(); !ok {
			ctx, _ = context.WithDeadline(ctx, time.Now().Add(opt.timeout))
		}
		return handler(ctx, req)
	}
}

