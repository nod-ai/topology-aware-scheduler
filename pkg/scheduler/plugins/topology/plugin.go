package algorithm

import (
    "context"
    "fmt"
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/kubernetes/pkg/scheduler/framework"
)

type TopologySchedulerPlugin struct {
    handle    framework.Handle
    scheduler *TopologyScheduler
}

const (
    Name = "topology-aware-scheduler"
)

var _ framework.FilterPlugin = &TopologySchedulerPlugin{}
var _ framework.ScorePlugin = &TopologySchedulerPlugin{}

func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
    cache := NewTopologyCache(NewNodeCache())
    scheduler := NewTopologyScheduler(cache)
    
    return &TopologySchedulerPlugin{
        handle:    h,
        scheduler: scheduler,
    }, nil
}

func (tp *TopologySchedulerPlugin) Name() string {
    return Name
}

func (tp *TopologySchedulerPlugin) Filter(
    ctx context.Context,
    state *framework.CycleState,
    pod *v1.Pod,
    nodeInfo *framework.NodeInfo,
) *framework.Status {
    if nodeInfo.Node() == nil {
        return framework.NewStatus(framework.Error, "node not found")
    }

    gpuReq, err := tp.scheduler.getGPURequirements(pod)
    if err != nil {
        return framework.NewStatus(framework.Unschedulable, 
            fmt.Sprintf("failed to get GPU requirements: %v", err))
    }

    domain, err := tp.scheduler.cache.GetDomainForNode(nodeInfo.Node().Name)
    if err != nil {
        return framework.NewStatus(framework.Error, 
            fmt.Sprintf("failed to get domain: %v", err))
    }

    if !tp.scheduler.isDomainEligible(domain, gpuReq) {
        return framework.NewStatus(framework.Unschedulable,
            "node's domain does not meet GPU requirements")
    }

    return framework.NewStatus(framework.Success, "")
}

func (tp *TopologySchedulerPlugin) Score(
    ctx context.Context,
    state *framework.CycleState,
    pod *v1.Pod,
    nodeName string,
) (int64, *framework.Status) {
    nodeInfo, err := tp.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
    if err != nil {
        return 0, framework.NewStatus(framework.Error,
            fmt.Sprintf("failed to get node info: %v", err))
    }

    gpuReq, err := tp.scheduler.getGPURequirements(pod)
    if err != nil {
        return 0, framework.NewStatus(framework.Error,
            fmt.Sprintf("failed to get GPU requirements: %v", err))
    }

    domain, err := tp.scheduler.cache.GetDomainForNode(nodeName)
    if err != nil {
        return 0, framework.NewStatus(framework.Error,
            fmt.Sprintf("failed to get domain: %v", err))
    }

    score := tp.scheduler.calculateDomainScore(domain, gpuReq)
    return int64(score * 100), framework.NewStatus(framework.Success,
