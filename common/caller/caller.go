package caller

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"github.com/openzipkin/zipkin-go"
)

const (
	metaKey4MPService     = "mp-service"
	metaKey4RealIP        = "x-real-ip"
	metaKey4XForwardedFor = "x-forwarded-for"
	metaKey4UserAgent     = "user-agent"
)

func GetMPService(ctx context.Context) string {
	return getMetaFromMetadataContext(ctx, metaKey4MPService)
}

func GetRealAddr(ctx context.Context) string {
	return getMetaFromMetadataContext(ctx, metaKey4RealIP)
}

func getMetaFromMetadataContext(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	rips := md.Get(key)
	if len(rips) == 0 {
		return ""
	}

	return rips[0]
}

func GetForwardedFor(ctx context.Context) string {
	return getMetaFromMetadataContext(ctx, metaKey4XForwardedFor)
}

func GetUserAgent(ctx context.Context) string {
	return getMetaFromMetadataContext(ctx, metaKey4UserAgent)
}

func GetPeerAddr(ctx context.Context) string {
	var addr string
	if pr, ok := peer.FromContext(ctx); ok {
		if tcpAddr, ok := pr.Addr.(*net.TCPAddr); ok {
			addr = tcpAddr.IP.String()
		} else {
			addr = pr.Addr.String()
		}
	}
	return addr
}

func SetXB3TraceID(ctx context.Context) context.Context {
	md := metadata.Pairs("X-B3-Traceid", GetTraceIDFromContext(ctx))
	ctx = metadata.NewIncomingContext(ctx, md)
	return ctx
}

func SplitMethodName(fullMethodName string) (service, method string) {
	service = "unknown"
	method = "method"

	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		service = fullMethodName[:i]
		method = fullMethodName[i+1:]
	}
	return
}

func GetTraceIDFromContext(ctx context.Context) (traceID string) {
	if span := zipkin.SpanFromContext(ctx); span != nil {
		traceID = span.Context().TraceID.String()
	}

	return
}
