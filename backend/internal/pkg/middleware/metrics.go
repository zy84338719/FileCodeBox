// Package middleware 提供 HTTP 中间件。
//
// 本文件实现 Prometheus 指标采集中间件，暴露 HTTP RED 指标：
//   - RequestsTotal（Counter）：请求总数，按 method/path/status 打标
//   - RequestDurationSeconds（Histogram）：请求延迟分布
//   - RequestsInFlight（Gauge）：当前处理中的请求数
//
// 指标通过 /metrics 端点暴露（由 bootstrap customizedRegister 注册 promhttp.Handler）。
package middleware

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics 指标容器。
type Metrics struct {
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight *prometheus.GaugeVec
}

// metricsInstance 单例（由 NewMetrics 注册后持有，供 Handler 访问）。
var metricsInstance *Metrics

// standardBuckets 适合 HTTP API 的延迟桶（秒）。
var standardBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// NewMetrics 创建并向默认注册表注册指标。
// 调用一次即可；重复调用会 panic（prometheus 已注册）。
func NewMetrics() *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		}, []string{"method", "path", "status"}),
		requestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: standardBuckets,
		}, []string{"method", "path"}),
		requestsInFlight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed.",
		}, []string{"method"}),
	}
	prometheus.MustRegister(m.requestsTotal, m.requestDuration, m.requestsInFlight)
	metricsInstance = m
	return m
}

// GetMetrics 返回已注册的指标实例（未初始化返回 nil）。
func GetMetrics() *Metrics {
	return metricsInstance
}

// normalizePath 收敛动态路径，避免高基数标签（如 /share/:code 而非 /share/ABC123）。
// 这里做轻量处理：将路径分段中疑似 ID/code 的部分替换为占位符。
func normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	// 简单规则：分段全为字母数字混合且长度>=8 视为动态 ID，替换为 :id
	// 这能覆盖分享码、UUID、自增 ID 等常见情况，同时保留静态路由。
	out := make([]byte, 0, len(path))
	segStart := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			seg := path[segStart:i]
			if isDynamicSegment(seg) {
				out = append(out, '/', ':', 'i', 'd')
			} else if len(seg) > 0 {
				out = append(out, '/')
				out = append(out, seg...)
			}
			segStart = i + 1
		}
	}
	if len(out) == 0 {
		return path
	}
	return string(out)
}

// isDynamicSegment 判断分段是否疑似动态 ID（高基数）。
func isDynamicSegment(s string) bool {
	if len(s) < 6 {
		return false
	}
	// 纯数字（自增 ID）
	isNum := true
	for _, c := range s {
		if c < '0' || c > '9' {
			isNum = false
			break
		}
	}
	if isNum {
		return true
	}
	// 长度 >= 8 的字母数字混合（分享码/UUID 等）
	hasLetter, hasDigit := false, false
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9':
			hasDigit = true
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			hasLetter = true
		default:
			return false // 含特殊字符，视为静态路径
		}
	}
	return hasLetter && hasDigit && len(s) >= 8
}

// MetricsMiddleware 采集 HTTP RED 指标的中间件。
// 必须在 NewMetrics() 之后、路由注册之前通过 h.Use 注册为全局中间件。
func MetricsMiddleware() app.HandlerFunc {
	m := metricsInstance
	return func(ctx context.Context, c *app.RequestContext) {
		if m == nil {
			c.Next(ctx)
			return
		}
		method := string(c.Method())
		// 在 Next 之前读取原始请求路径，避免被下游 handler（如 c.File）改写污染标签
		path := normalizePath(string(c.Request.URI().Path()))
		m.requestsInFlight.WithLabelValues(method).Inc()
		start := time.Now()

		c.Next(ctx)

		latency := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response.StatusCode())

		m.requestsTotal.WithLabelValues(method, path, status).Inc()
		m.requestDuration.WithLabelValues(method, path).Observe(latency)
		m.requestsInFlight.WithLabelValues(method).Dec()
	}
}
