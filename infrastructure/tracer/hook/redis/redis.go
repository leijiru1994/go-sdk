package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus"
)

type ctxKey struct{}

var startedKey = ctxKey{}

type TracingHook struct {
	tracer *zipkin.Tracer
	instance string
}

var _ redis.Hook = TracingHook{}

var StorageCmdLatencyMetric *prometheus.HistogramVec

func init() {
	StorageCmdLatencyMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "storage_server_cmd_latency_stats",
			Help: "Histogram of response latency (seconds) of storage that had been application-level handled by the server.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"storage_instance", "storage_service", "storage_method"})
	prometheus.MustRegister(StorageCmdLatencyMetric)
}

// NewHook creates a new go-redis hook instance and that will collect spans using the provided tracer.
func NewHook(tracer *zipkin.Tracer, instance string) redis.Hook {
	return &TracingHook{
		tracer: tracer,
		instance: instance,
	}
}

func (hook TracingHook) createSpan(ctx context.Context, operationName string) (zipkin.Span, context.Context) {
	span := zipkin.SpanFromContext(ctx)
	if span != nil {
		childSpan := hook.tracer.StartSpan(
			operationName,
			zipkin.Parent(span.Context()))
		return childSpan, zipkin.NewContext(ctx, childSpan)
	}

	return hook.tracer.StartSpanFromContext(
		ctx,
		operationName)
}

func (hook TracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, ctx := hook.createSpan(ctx, fmt.Sprintf("redis_%v", cmd.FullName()))
	span.Tag("db.type", "redis")
	span.Tag("db.cmd", cmd.String())

	// add start time in context
	ctx = context.WithValue(ctx, startedKey, time.Now())

	return ctx, nil
}

func (hook TracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := zipkin.SpanFromContext(ctx)
	defer func() {
		go span.Finish()
	}()

	if err := cmd.Err(); err != nil {
		recordError(ctx, "db.error", span, err)
	}

	if d, ok := ctx.Value(startedKey).(time.Time); ok {
		StorageCmdLatencyMetric.WithLabelValues(
			hook.instance,
			"redis",
			cmd.Name(),
		).Observe(time.Since(d).Seconds())
	}

	return nil
}

func (hook TracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	span, ctx := hook.createSpan(ctx, "pipeline")
	span.Tag("db.type", "redis")
	span.Tag("db.redis.cmd_nums", strconv.Itoa(len(cmds)))
	return ctx, nil
}

func (hook TracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := zipkin.SpanFromContext(ctx)
	defer func() {
		go span.Finish()
	}()

	for i, cmd := range cmds {
		if err := cmd.Err(); err != nil {
			recordError(ctx, "redis.error"+strconv.Itoa(i), span, err)
		}
	}
	return nil
}

func recordError(ctx context.Context, errorTag string, span zipkin.Span, err error) {
	if err != redis.Nil {
		span.Tag(errorTag, err.Error())
	}
}
