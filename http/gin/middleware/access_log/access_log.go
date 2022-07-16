package access_log

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/leijiru1994/go-sdk/common/caller"
	"github.com/leijiru1994/go-sdk/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type statusRecorder struct {
	gin.ResponseWriter

	status int
	data   interface{}
}

func (w *statusRecorder) Write(p []byte) (int, error) {
	m := map[string]interface{}{}
	_ = json.Unmarshal(p, &m)
	w.data = m

	return w.ResponseWriter.Write(p)
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// WrapWriter interface
func (w *statusRecorder) WrappedWriter() http.ResponseWriter {
	return w.ResponseWriter
}

// CloseNotify implements the http.CloseNotify interface.
func (w *statusRecorder) CloseNotify() <-chan bool {
	if d ,ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return d.CloseNotify()
	}

	return make(chan bool)
}

func WithAccessLog(opts ...Option) gin.HandlerFunc {
	opt := evaluateOptions(opts)
	return func(ctx *gin.Context) {
		if opt == nil {
			ctx.Next()
		} else {
			writer := &statusRecorder{ctx.Writer, 200, ""}
			startedAt := time.Now()
			var bodyBytes []byte
			if ctx.Request.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(ctx.Request.Body)
			}
			ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

			mm := make(map[string]interface{})
			bb := binding.Default(ctx.Request.Method, ctx.Request.Header.Get("Content-Type"))
			switch bb {
			case binding.JSON:
				_ = json.Unmarshal(bodyBytes, &mm)
			default:
			}

			defer func() {
				status := ctx.Writer.Status()
				select {
				case <-ctx.Request.Context().Done():
					// NOTE: if client is keep-alive, 499 may not appeared
					// Refer: https://github.com/golang/go/issues/13165
					status = 499
				default:
					// Nothing to do
				}

				logInfo := map[string]interface{}{
					"version":              ctx.Request.Header.Get("version"),
					"host":                 strings.Split(ctx.Request.Host, ":")[0],
					"client_ip":            strings.Split(ctx.Request.RemoteAddr, ":")[0],
					"request_method":       ctx.Request.Method,
					"path":                 ctx.Request.URL.Path,
					"request_url":          ctx.Request.RequestURI,
					"status":               status,
					"http_user_agent":      ctx.Request.UserAgent(),
					"request_time":         time.Since(startedAt).Seconds(),
					"http_x_forwarded_for": GetIP(ctx.Request),
					"request_params":       mm,
					"response":             writer.data,
					"trace_id":             caller.GetTraceIDFromContext(ctx.Request.Context()),
					"user_id":              util.UserIDFromCtx(ctx),
				}
				opt.logger.Log().Interface("message", logInfo).Int64("time", time.Now().Unix()).Send()
			}()

			ctx.Writer = writer
            ctx.Next()
		}
	}
}

// GetIP 获取连接ip
func GetIP(r *http.Request) string {
	// 先从HTTP_X_CLUSTER_CLIENT_IP获取
	ip := r.Header.Get("HTTP_X_CLUSTER_CLIENT_IP")
	if ip == "" {
		ip = r.Header.Get("HTTP_CLIENT_IP")
		if ip == "" {
			ip = r.Header.Get("HTTP_X_FORWARDED_FOR")
			if ip == "" {
				ip = r.Header.Get("X-FORWARDED-FOR")
				if ip == "" {
					ip = strings.Split(r.RemoteAddr, ":")[0]
				}
			}
		}
	}
	return strings.Split(ip, ",")[0]
}
