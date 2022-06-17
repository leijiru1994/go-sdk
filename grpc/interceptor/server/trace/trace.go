package trace

import (
	"context"
	"os"
	"strconv"

	common "github.com/leijiru1994/go-sdk/common/model"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// TODO: 待更新到V2
// UnaryServerInterceptor returns a new unary server interceptor for OpenTracing.
func UnaryServerInterceptor(tracer *zipkin.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if tracer == nil {
			return handler(ctx, req)
		}

		spanName := info.FullMethod
		newCtx, serverSpan := newServerSpanFromInbound(ctx, tracer, spanName)
		resp, err := handler(newCtx, req)
		finishSpan(ctx, serverSpan, resp, err)
		return resp, err
	}
}

func newServerSpanFromInbound(ctx context.Context, tracer *zipkin.Tracer, spanName string) (context.Context, zipkin.Span) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	sc := tracer.Extract(b3.ExtractGRPC(&md))
	span, newCtx := tracer.StartSpanFromContext(
		ctx,
		spanName,
		zipkin.Kind(model.Server),
		zipkin.Parent(sc),
		zipkin.RemoteEndpoint(remoteEndpointFromContext(ctx, os.Getenv("PROJECT_NAME"))))
	if !zipkin.IsNoop(span) {
		// TODO: we will add some default tags in each project
	}

	return newCtx, span
}

func finishSpan(ctx context.Context, span zipkin.Span, resp interface{}, err error) {
	code := "-1"
	if err != nil {
		code = err.Error()
	} else if resp != nil {
		if r, ok := resp.(common.CommonResponse); ok {
			code = strconv.FormatUint(uint64(r.GetCode()), 10)
		}
	}

	span.Tag("code", code)

	// 这里不需要关心是否上传成功，后续用kafka做投递
	go span.Finish()
}

// NOTE: will be remove name on the time
func remoteEndpointFromContext(ctx context.Context, name string) *model.Endpoint {
	remoteAddr := ""

	p, ok := peer.FromContext(ctx)
	if ok {
		remoteAddr = p.Addr.String()
	}

	ep, _ := zipkin.NewEndpoint(name, remoteAddr)
	return ep
}
