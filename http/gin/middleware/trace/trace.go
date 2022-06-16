package trace

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"google.golang.org/grpc/metadata"
)

func WithTrace(opts ...Option) gin.HandlerFunc {
	opt := evaluateOptions(opts)
	return func(ctx *gin.Context) {
		if opt.tracer == nil {
			ctx.Next()
		} else {
			//clientSpan, newCtx := opt.newClientSpanFromContext(ctx, opt.tracer, method)
			//ctx.Next()
			//clientSpan.Tag("grpc.target", cc.Target())
			//finishClientSpan(clientSpan, err)
		}
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