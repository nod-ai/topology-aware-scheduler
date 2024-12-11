go
package metrics

import (
    "time"
    "sync"
    "github.com/prometheus/client_golang/prometheus"
)

type MetricTimestamp struct {
    startTime time.Time
    operation string
}

type MetricsCache struct {
    mu     sync.RWMutex
    values map[string]float64
}

func NewMetricsCache() *MetricsCache {
    return &MetricsCache{
        values: make(map[string]float64),
    }
}

func StartMetricTimer(operation string) MetricTimestamp {
    return MetricTimestamp{
        startTime: time.Now(),
        operation: operation,
    }
}

func (mt *MetricTimestamp) RecordDuration(histogram *prometheus.HistogramVec, labels ...string) {
    duration := time.Since(mt.startTime).Seconds()
    histogram.WithLabelValues(labels...).Observe(duration)
}

func CalculateEMA(current, new float64, alpha float64) float64 {
    return alpha*new + (1-alpha)*current
}

func UpdateHistogram(histogram *prometheus.HistogramVec, value float64, labels ...string) {
    if value < 0 {
        value = 0
    }
    histogram.WithLabelValues(labels...).Observe(value)
}

func RecordError(errorCounter *prometheus.CounterVec, errType string, err error) {
    if err != nil {
        errorCounter.WithLabelValues(errType).Inc()
    }
}

func (mc *MetricsCache) UpdateValue(key string, value float64) {
    mc.mu.Lock()
    defer mc.mu.Unlock()
    mc.values[key] = value
}

func (mc *MetricsCache) GetValue(key string) (float64, bool) {
    mc.mu.RLock()
    defer mc.mu.RUnlock()
    value, exists := mc.values[key]
    return value, exists
}
