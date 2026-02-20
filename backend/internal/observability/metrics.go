package observability

import (
	"math"
	"sort"
	"sync"
	"time"
)

const (
	maxLatencySamples = 5000
	qpsWindowSeconds  = 60
)

type metricBucket struct {
	Count          uint64
	ErrorCount     uint64
	TotalLatencyMs float64
	LatencySamples []float64
}

type collector struct {
	mu sync.Mutex

	startedAt time.Time

	httpSummary metricBucket
	routes      map[string]*metricBucket
	downstreams map[string]*metricBucket

	requestTimestamps []int64
}

var globalCollector = &collector{
	startedAt:   time.Now(),
	routes:      make(map[string]*metricBucket),
	downstreams: make(map[string]*metricBucket),
}

type BucketSnapshot struct {
	Count        uint64  `json:"count"`
	ErrorCount   uint64  `json:"error_count"`
	ErrorRate    float64 `json:"error_rate"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
	P95LatencyMs float64 `json:"p95_latency_ms"`
	P99LatencyMs float64 `json:"p99_latency_ms"`
}

type SummarySnapshot struct {
	TotalRequests uint64  `json:"total_requests"`
	TotalErrors   uint64  `json:"total_errors"`
	ErrorRate     float64 `json:"error_rate"`
	QPSAvg        float64 `json:"qps_avg"`
	QPS1m         float64 `json:"qps_1m"`
	AvgLatencyMs  float64 `json:"avg_latency_ms"`
	P95LatencyMs  float64 `json:"p95_latency_ms"`
	P99LatencyMs  float64 `json:"p99_latency_ms"`
}

type Snapshot struct {
	Timestamp  string                    `json:"timestamp"`
	UptimeSec  int64                     `json:"uptime_sec"`
	Summary    SummarySnapshot           `json:"summary"`
	Routes     map[string]BucketSnapshot `json:"routes"`
	Downstream map[string]BucketSnapshot `json:"downstream"`
}

func addSample(bucket *metricBucket, latencyMs float64, isError bool) {
	bucket.Count++
	if isError {
		bucket.ErrorCount++
	}
	bucket.TotalLatencyMs += latencyMs
	bucket.LatencySamples = append(bucket.LatencySamples, latencyMs)
	if len(bucket.LatencySamples) > maxLatencySamples {
		overflow := len(bucket.LatencySamples) - maxLatencySamples
		bucket.LatencySamples = bucket.LatencySamples[overflow:]
	}
}

func formatBucket(bucket *metricBucket) BucketSnapshot {
	if bucket == nil || bucket.Count == 0 {
		return BucketSnapshot{}
	}

	avg := bucket.TotalLatencyMs / float64(bucket.Count)
	p95 := percentile(bucket.LatencySamples, 95)
	p99 := percentile(bucket.LatencySamples, 99)
	errorRate := float64(bucket.ErrorCount) / float64(bucket.Count)

	return BucketSnapshot{
		Count:        bucket.Count,
		ErrorCount:   bucket.ErrorCount,
		ErrorRate:    round(errorRate, 4),
		AvgLatencyMs: round(avg, 2),
		P95LatencyMs: round(p95, 2),
		P99LatencyMs: round(p99, 2),
	}
}

func percentile(samples []float64, p int) float64 {
	if len(samples) == 0 {
		return 0
	}
	copySamples := make([]float64, len(samples))
	copy(copySamples, samples)
	sort.Float64s(copySamples)

	rank := int(math.Ceil((float64(p) / 100) * float64(len(copySamples))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(copySamples) {
		rank = len(copySamples)
	}

	return copySamples[rank-1]
}

func round(value float64, digits int) float64 {
	pow := math.Pow(10, float64(digits))
	return math.Round(value*pow) / pow
}

func errorStatus(statusCode int) bool {
	return statusCode >= 400
}

func routeKey(method, route string) string {
	return method + " " + route
}

func downstreamKey(service, operation string) string {
	return service + ":" + operation
}

func (c *collector) purgeQPSWindow(now int64) {
	cutoff := now - qpsWindowSeconds + 1
	idx := 0
	for idx < len(c.requestTimestamps) && c.requestTimestamps[idx] < cutoff {
		idx++
	}
	if idx > 0 {
		c.requestTimestamps = c.requestTimestamps[idx:]
	}
}

// RecordHTTPRequest 记录 HTTP 层请求指标。
func RecordHTTPRequest(method, route string, statusCode int, latency time.Duration) {
	if route == "" {
		route = "UNKNOWN"
	}

	now := time.Now().Unix()
	latencyMs := float64(latency.Milliseconds())

	globalCollector.mu.Lock()
	defer globalCollector.mu.Unlock()

	isError := errorStatus(statusCode)
	addSample(&globalCollector.httpSummary, latencyMs, isError)

	key := routeKey(method, route)
	bucket := globalCollector.routes[key]
	if bucket == nil {
		bucket = &metricBucket{}
		globalCollector.routes[key] = bucket
	}
	addSample(bucket, latencyMs, isError)

	globalCollector.requestTimestamps = append(globalCollector.requestTimestamps, now)
	globalCollector.purgeQPSWindow(now)
}

// RecordDownstream 记录下游服务调用指标（如 backend -> agent）。
func RecordDownstream(service, operation string, statusCode int, latency time.Duration) {
	if service == "" {
		service = "unknown"
	}
	if operation == "" {
		operation = "unknown"
	}

	latencyMs := float64(latency.Milliseconds())
	isError := statusCode <= 0 || errorStatus(statusCode)

	globalCollector.mu.Lock()
	defer globalCollector.mu.Unlock()

	key := downstreamKey(service, operation)
	bucket := globalCollector.downstreams[key]
	if bucket == nil {
		bucket = &metricBucket{}
		globalCollector.downstreams[key] = bucket
	}
	addSample(bucket, latencyMs, isError)
}

// SnapshotMetrics 获取当前指标快照。
func SnapshotMetrics() Snapshot {
	globalCollector.mu.Lock()
	defer globalCollector.mu.Unlock()

	now := time.Now()
	uptimeSec := int64(now.Sub(globalCollector.startedAt).Seconds())
	if uptimeSec < 1 {
		uptimeSec = 1
	}

	globalCollector.purgeQPSWindow(now.Unix())

	summary := formatBucket(&globalCollector.httpSummary)
	qpsAvg := float64(globalCollector.httpSummary.Count) / float64(uptimeSec)
	qps1m := float64(len(globalCollector.requestTimestamps)) / qpsWindowSeconds

	routes := make(map[string]BucketSnapshot, len(globalCollector.routes))
	for key, bucket := range globalCollector.routes {
		routes[key] = formatBucket(bucket)
	}

	downstream := make(map[string]BucketSnapshot, len(globalCollector.downstreams))
	for key, bucket := range globalCollector.downstreams {
		downstream[key] = formatBucket(bucket)
	}

	return Snapshot{
		Timestamp: now.Format(time.RFC3339),
		UptimeSec: uptimeSec,
		Summary: SummarySnapshot{
			TotalRequests: globalCollector.httpSummary.Count,
			TotalErrors:   globalCollector.httpSummary.ErrorCount,
			ErrorRate:     summary.ErrorRate,
			QPSAvg:        round(qpsAvg, 4),
			QPS1m:         round(qps1m, 4),
			AvgLatencyMs:  summary.AvgLatencyMs,
			P95LatencyMs:  summary.P95LatencyMs,
			P99LatencyMs:  summary.P99LatencyMs,
		},
		Routes:     routes,
		Downstream: downstream,
	}
}
