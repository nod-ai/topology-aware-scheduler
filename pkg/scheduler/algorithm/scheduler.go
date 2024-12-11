package algorithm

import (
    "context"
    "fmt"
    "sort"
    "sync"
    "time"
    v1 "k8s.io/api/core/v1"
)

type TopologyScheduler struct {
    sync.RWMutex
    cache            *TopologyCache
    scoreWeights     TopologyScore
    domains          map[string]*Domain
    spineConnections map[string][]string
    metrics          *MetricsCollector
    monitor          *DomainMonitor
}

func NewTopologyScheduler(cache *TopologyCache) *TopologyScheduler {
    ts := &TopologyScheduler{
        cache: cache,
        scoreWeights: TopologyScore{
            ResourceAvailability: 0.4,
            TopologyAlignment:    0.3,
            DomainUtilization:    0.2,
            HistoricalPerf:       0.1,
        },
        domains:          make(map[string]*Domain),
        spineConnections: make(map[string][]string),
        metrics:          NewMetricsCollector(),
    }
    ts.monitor = NewDomainMonitor(ts)
    return ts
}

func (ts *TopologyScheduler) Schedule(ctx context.Context, pod *v1.Pod) (*v1.Node, error) {
    startTime := time.Now()
    defer func() {
        ts.metrics.ObserveSchedulingLatency(time.Since(startTime))
    }()

    gpuReq, err := ts.getGPURequirements(pod)
    if err != nil {
        ts.metrics.IncSchedulingError("invalid_gpu_requirements")
        return nil, fmt.Errorf("failed to get GPU requirements: %v", err)
    }

    strategy := ts.getPlacementStrategy(gpuReq)
    var result *PlacementResult
    
    switch strategy {
    case SingleDomain:
        result, err = ts.placePodSingleDomain(ctx, pod, gpuReq)
    case CompleteDomain:
        result, err = ts.placeCompleteDomain(ctx, pod, gpuReq)
    case AdjacentDomains:
        result, err = ts.placePodAdjacentDomains(ctx, pod, gpuReq)
    case MultipleDomains:
        result, err = ts.placePodMultipleDomains(ctx, pod, gpuReq)
    default:
        ts.metrics.IncSchedulingError("invalid_strategy")
        return nil, fmt.Errorf("unsupported placement strategy")
    }

    if err != nil {
        ts.metrics.IncSchedulingError(fmt.Sprintf("placement_%s", strategy))
        return nil, err
    }

    ts.metrics.ObservePlacementResult(result)
    ts.updateDomainState(result)

    return result.Nodes[0], nil
}

func (ts *TopologyScheduler) getPlacementStrategy(gpuReq *GPURequirements) PlacementStrategy {
    switch {
    case gpuReq.NodesNeeded <= 2:
        return SingleDomain
    case gpuReq.NodesNeeded == 4:
        return CompleteDomain
    case gpuReq.NodesNeeded <= 8:
        return AdjacentDomains
    default:
        return MultipleDomains
    }
}

func (ts *TopologyScheduler) findCompleteFreeDomains() []*Domain {
    var freeDomains []*Domain
    for _, domain := range ts.domains {
        if domain.UsedGPUs == 0 {
            freeDomains = append(freeDomains, domain)
        }
    }
    return freeDomains
}

func (ts *TopologyScheduler) selectNodesAcrossDomains(domains []*Domain, gpuReq *GPURequirements) ([]*v1.Node, error) {
    var selectedNodes []*v1.Node
    remainingNodes := gpuReq.NodesNeeded

    for _, domain := range domains {
        availableNodes := ts.getAvailableNodes(domain)
        if len(availableNodes) == 0 {
            continue
        }

        nodesFromDomain := min(remainingNodes, len(availableNodes))
        selectedNodes = append(selectedNodes, availableNodes[:nodesFromDomain]...)
        remainingNodes -= nodesFromDomain

        if remainingNodes == 0 {
            break
        }
    }

    if remainingNodes > 0 {
        return nil, fmt.Errorf("insufficient nodes across domains")
    }

    return selectedNodes, nil
}

func (ts *TopologyScheduler) updateNodeState(node *v1.Node) error {
    ts.Lock()
    defer ts.Unlock()

    domain := ts.getDomainForNode(node)
    if domain == nil {
        return fmt.Errorf("node %s not found in any domain", node.Name)
    }

    gpus := ts.getUsedGPUs(node)
    domain.UsedGPUs += gpus
    return nil
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
