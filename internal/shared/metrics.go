package shared

import (
	"context"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type AppMetrics struct {
	requestDuration    *prometheus.HistogramVec
	requestTotal       *prometheus.CounterVec
	activeConnections  prometheus.Gauge
	memoryUsage        prometheus.Gauge
	cpuUsage           prometheus.Gauge
	goroutines         prometheus.Gauge
	todoOperations     *prometheus.CounterVec
	userOperations     *prometheus.CounterVec
	databaseOperations *prometheus.CounterVec
	rateLimitHits      *prometheus.CounterVec
	rateLimitAllowed   *prometheus.CounterVec
	cacheHits          *prometheus.CounterVec
	cacheMisses        *prometheus.CounterVec
}

func NewAppMetrics(registry prometheus.Registerer) *AppMetrics {
	metrics := &AppMetrics{
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		requestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		activeConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
		memoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
		),
		cpuUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),
		goroutines: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "goroutines_total",
				Help: "Number of goroutines",
			},
		),
		todoOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "todo_operations_total",
				Help: "Total number of todo operations",
			},
			[]string{"operation"},
		),
		userOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "user_operations_total",
				Help: "Total number of user operations",
			},
			[]string{"operation"},
		),
		databaseOperations: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_operations_total",
				Help: "Total number of database operations",
			},
			[]string{"operation", "table"},
		),
		rateLimitHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"path", "key_type"},
		),
		rateLimitAllowed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rate_limit_allowed_total",
				Help: "Total number of requests allowed by rate limiter",
			},
			[]string{"path", "key_type"},
		),
		cacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
			[]string{"path"},
		),
		cacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
			[]string{"path"},
		),
	}

	registry.MustRegister(
		metrics.requestDuration,
		metrics.requestTotal,
		metrics.activeConnections,
		metrics.memoryUsage,
		metrics.cpuUsage,
		metrics.goroutines,
		metrics.todoOperations,
		metrics.userOperations,
		metrics.databaseOperations,
		metrics.rateLimitHits,
		metrics.rateLimitAllowed,
		metrics.cacheHits,
		metrics.cacheMisses,
	)

	return metrics
}

func (m *AppMetrics) RecordRequest(ctx context.Context, method, path, status string, duration time.Duration) {
	m.requestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
	m.requestTotal.WithLabelValues(method, path, status).Inc()
}

func (m *AppMetrics) IncrementActiveConnections(ctx context.Context) {
	m.activeConnections.Inc()
}

func (m *AppMetrics) DecrementActiveConnections(ctx context.Context) {
	m.activeConnections.Dec()
}

func (m *AppMetrics) RecordTodoOperation(ctx context.Context, operation string) {
	m.todoOperations.WithLabelValues(operation).Inc()
}

func (m *AppMetrics) RecordUserOperation(ctx context.Context, operation string) {
	m.userOperations.WithLabelValues(operation).Inc()
}

func (m *AppMetrics) RecordDatabaseOperation(ctx context.Context, operation, table string) {
	m.databaseOperations.WithLabelValues(operation, table).Inc()
}

func (m *AppMetrics) RecordRateLimitHit(ctx context.Context, path, keyType string) {
	m.rateLimitHits.WithLabelValues(path, keyType).Inc()
}

func (m *AppMetrics) RecordRateLimitAllowed(ctx context.Context, path, keyType string) {
	m.rateLimitAllowed.WithLabelValues(path, keyType).Inc()
}

func (m *AppMetrics) RecordCacheHit(ctx context.Context, path string) {
	m.cacheHits.WithLabelValues(path).Inc()
}

func (m *AppMetrics) RecordCacheMiss(ctx context.Context, path string) {
	m.cacheMisses.WithLabelValues(path).Inc()
}

func (m *AppMetrics) StartSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Memory metrics
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)
				m.memoryUsage.Set(float64(memStats.Alloc))

				// Goroutines count
				m.goroutines.Set(float64(runtime.NumGoroutine()))

			case <-ctx.Done():
				return
			}
		}
	}()
}
