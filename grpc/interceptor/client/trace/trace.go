package trace

import (
	"context"
	"time"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a new unary client interceptor for OpenTracing.
func UnaryServerInterceptor(opts ...Option) grpc.UnaryClientInterceptor {
	opt := evaluateOptions(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if opt.tracer == nil {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		clientSpan, newCtx := opt.newClientSpanFromContext(ctx, opt.tracer, method)
		err := invoker(newCtx, method, req, reply, cc, opts...)
		clientSpan.Tag("grpc.target", cc.Target())
		finishClientSpan(clientSpan, err)

		return err
	}
}

func (opts *options) newClientSpanFromContext(ctx context.Context, tracer *zipkin.Tracer, fullMethodName string) (zipkin.Span, context.Context) {
	sc := zipkin.SpanFromContext(ctx)
	spanOpts := []zipkin.SpanOption{
		zipkin.StartTime(time.Now()),
		zipkin.Kind(model.Client),
	}
	if sc != nil {
		spanOpts= append(spanOpts, zipkin.Parent(sc.Context()))
	}

	span := tracer.StartSpan(fullMethodName, spanOpts...)
	span.Tag("grpc.client", opts.client)

	md := &metadata.MD{}
	_ = b3.InjectGRPC(md)(span.Context())
	md.Set("mp-service", opts.client)
	return span, metadata.NewOutgoingContext(ctx, *md)
}

func finishClientSpan(clientSpan zipkin.Span, err error) {
	if err != nil {
		clientSpan.Tag("grpc.error", err.Error())
	}
	clientSpan.Finish()
}
