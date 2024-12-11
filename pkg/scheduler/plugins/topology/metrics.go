go
package algorithm

import (
    "sync"
    "time"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricsCollector struct {
    mu sync.RWMutex

    // Scheduling metrics
    schedulingLatency *prometheus.HistogramVec
    schedulingErrors *prometheus.CounterVec
    schedulingAttempts *prometheus.CounterVec
    schedulingSuccess *prometheus.CounterVec

    // Domain metrics
    domainUtilization *prometheus.GaugeVec
    gpuUtilization *prometheus.GaugeVec
    domainFragmentation *prometheus.GaugeVec

    // Node metrics
    nodeGPUAllocation *prometheus.GaugeVec
    nodeHealthStatus *prometheus.GaugeVec

    // Placement metrics
    placementDecisions *prometheus.CounterVec
    placementScores *prometheus.HistogramVec
}

func NewMetricsCollector() *MetricsCollector {
    return &MetricsCollector{
        schedulingLatency: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "topology_scheduler_latency_seconds",
                Help: "Time taken for scheduling decisions",
                Buckets: prometheus.ExponentialBuckets(0.001, 2, 15),
            },
            []string{"strategy"},
        ),

        schedulingErrors: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "topology_scheduler_errors_total",
                Help: "Total number of scheduling errors by type",
            },
            []string{"type"},
        ),

        schedulingAttempts: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "topology_scheduler_attempts_total",
                Help: "Number of scheduling attempts",
            },
            []string{"strategy"},
        ),

        schedulingSuccess: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "topology_scheduler_success_total",
                Help: "Number of successful schedules",
            },
            []string{"strategy"},
        ),

        domainUtilization: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "topology_domain_utilization_ratio",
                Help: "Current utilization ratio of domains",
            },
            []string{"domain"},
        ),

        gpuUtilization: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "topology_gpu_utilization_ratio",
                Help: "GPU utilization ratio by domain",
            },
            []string{"domain"},
        ),

        domainFragmentation: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "topology_domain_fragmentation_ratio",
                Help: "Fragmentation ratio of domains",
            },
            []string{"domain"},
        ),

        nodeGPUAllocation: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "topology_node_gpu_allocated",
                Help: "Number of GPUs allocated per node",
            },
            []string{"node", "domain"},
        ),

        nodeHealthStatus: promauto.NewGaugeVec(
            prometheus.GaugeOpts{
                Name: "topology_node_health_status",
                Help: "Health status of nodes (1 for healthy, 0 for unhealthy)",
            },
            []string{"node", "domain"},
        ),

        placementDecisions: promauto.NewCounterVec(
            prometheus.CounterOpts{
                Name: "topology_placement_decisions_total",
                Help: "Number of placement decisions by type",
            },
            []string{"strategy", "result"},
        ),

        placementScores: promauto.NewHistogramVec(
            prometheus.HistogramOpts{
                Name: "topology_placement_scores",
                Help: "Distribution of placement scores",
                Buckets: prometheus.LinearBuckets(0, 0.1, 10),
            },
            []string{"strategy"},
        ),
    }
}

func (mc *MetricsCollector) ObserveSchedulingLatency(duration time.Duration, strategy string) {
    mc.schedulingLatency.WithLabelValues(strategy).Observe(duration.Seconds())
}

func (mc *MetricsCollector) IncSchedulingError(errorType string) {
    mc.schedulingErrors.WithLabelValues(errorType).Inc()
}

func (mc *MetricsCollector) ObservePlacementResult(result *PlacementResult) {
    if result == nil {
        return
    }
    
    mc.placementDecisions.WithLabelValues(string(result.Strategy), "success").Inc()
    mc.placementScores.WithLabelValues(string(result.Strategy)).Observe(result.Score)
}

func (mc *MetricsCollector) UpdateDomainMetrics(domain *Domain) {
    mc.mu.Lock()
    defer mc.mu.Unlock()

    utilization := float64(domain.UsedGPUs) / float64(domain.TotalGPUs)
    mc.domainUtilization.WithLabelValues(domain.Name).Set(utilization)
    
    gpuUtil := float64(domain.UsedGPUs) / float64(len(domain.Nodes)*4) // Assuming 4 GPUs per node
    mc.gpuUtilization.WithLabelValues(domain.Name).Set(gpuUtil)
    
    fragmentation := calculateFragmentation(domain)
    mc.domainFragmentation.WithLabelValues(domain.Name).Set(fragmentation)
}

func (mc *MetricsCollector) UpdateNodeMetrics(node string, domain string, gpuCount int, healthy bool) {
    mc.nodeGPUAllocation.WithLabelValues(node, domain).Set(float64(gpuCount))
    
    healthStatus := 0.0
    if healthy {
        healthStatus = 1.0
    }
    mc.nodeHealthStatus.WithLabelValues(node, domain).Set(healthStatus)
}

func calculateFragmentation(domain *Domain) float64 {
    if domain.TotalGPUs == 0 {
        return 0.0
    }
    
    // Simple fragmentation metric: ratio of partially used nodes to total nodes
    partialNodes := 0
    for _, node := range domain.Nodes {
        gpusUsed := domain.UsedGPUs // This should be per-node in real implementation
        if gpusUsed > 0 && gpusUsed < 4 { // Assuming 4 GPUs per node
            partialNodes++
        }
    }
    
    return float64(partialNodes) / float64(len(domain.Nodes))
}
