package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:      "http_request_latency",
		Help:      "The http request latency in seconds",
	}, []string{"method", "path", "status", "code"})
)

func init() {
	prometheus.MustRegister(httpRequestLatency)
}

func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		begin := time.Now()
		c.Next()
		var code string
		if d, exists := c.Get("code"); exists && d != nil {
			code = strconv.Itoa(d.(int))
		}
		duration := float64(time.Since(begin)) / float64(time.Second)
		httpRequestLatency.With(prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"status": strconv.Itoa(c.Writer.Status()),
			"code":   code,
		}).Observe(duration)
	}
}
