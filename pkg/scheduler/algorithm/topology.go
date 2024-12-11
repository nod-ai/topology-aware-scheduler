package scheduler

import (
    "sync"
    "k8s.io/api/core/v1"
)

type TopologyManager struct {
    mu sync.RWMutex
    domainManager *DomainManager
    nodeManager   *NodeManager
}

func NewTopologyManager() *TopologyManager {
    return &TopologyManager{
        domainManager: NewDomainManager(),
        nodeManager:   NewNodeManager(),
    }
}

func (tm *TopologyManager) UpdateNode(node *v1.Node) error {
    tm.mu.Lock()
    defer tm.mu.Unlock()

    nodeInfo := &Node{
        Name: node.Name,
        GPUs: parseGPUInfo(node),
        NetworkBandwidth: parseNetworkBandwidth(node),
    }

    domain := parseDomainInfo(node)
    if err := tm.domainManager.AddDomain(domain); err != nil {
        return err
    }

    return tm.nodeManager.UpdateNode(nodeInfo)
}

func (tm *TopologyManager) GetTopologyDistance(source, target string) (int, error) {
    sourceDomain, err := tm.domainManager.GetDomainByNode(source)
    if err != nil {
        return 0, err
    }

    targetDomain, err := tm.domainManager.GetDomainByNode(target)
    if err != nil {
        return 0, err
    }

    return calculateTopologyDistance(sourceDomain, targetDomain), nil
}
