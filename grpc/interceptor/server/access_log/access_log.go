package access_log

import (
	"context"
	"time"

	"go-sdk/common/caller"
	"go-sdk/ecode"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(opts ...Option) grpc.UnaryServerInterceptor {
	opt := evaluateOptions(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if opt == nil {
			return handler(ctx, req)
		}

		startedAt := time.Now()
		resp, err := handler(ctx, req)
		switch err.(type) {
		case ecode.Code:
			err = status.Error(codes.Code(err.(ecode.Code)), err.(ecode.Code).Message())
		case error:
			if status.Code(err) == codes.InvalidArgument {
				err = status.Error(codes.InvalidArgument, "参数错误")
			} else {
				err = status.Error(codes.Unknown, "非grpc status error")
			}
		}

		logInfo := map[string]interface{}{
			"mp_service":           caller.GetMPService(ctx),
			"client_ip":            caller.GetPeerAddr(ctx),
			"real_ip":              caller.GetRealAddr(ctx),
			"user_agent":           caller.GetUserAgent(ctx),
			"http_x_forwarded_for": caller.GetForwardedFor(ctx),
			"method":               info.FullMethod,
			"status":               int(status.Code(err)),
			"request_time":         startedAt.UnixNano() / 1e6,
			"duration":             time.Since(startedAt).Milliseconds(),
			"request_params":       req,
			"response":             resp,
			"trace_id":             caller.GetTraceIDFromContext(ctx),
		}
		opt.logger.Log().Interface("message", logInfo).Int64("time", time.Now().Unix()).Send()

		return resp, err
	}
}
