package algorithm

import (
    "fmt"
    "sync"
    "time"
    v1 "k8s.io/api/core/v1"
)

type NodeCache struct {
    sync.RWMutex
    nodes             map[string]*v1.Node
    gpuAllocations    map[string]int
    lastNodeUpdate    map[string]time.Time
    metrics           *MetricsCollector
}

func NewNodeCache() *NodeCache {
    return &NodeCache{
        nodes:          make(map[string]*v1.Node),
        gpuAllocations: make(map[string]int),
        lastNodeUpdate: make(map[string]time.Time),
        metrics:        NewMetricsCollector(),
    }
}

func (nc *NodeCache) AddNode(node *v1.Node) error {
    nc.Lock()
    defer nc.Unlock()

    if _, exists := nc.nodes[node.Name]; exists {
        return fmt.Errorf("node %s already exists", node.Name)
    }

    nc.nodes[node.Name] = node
    nc.gpuAllocations[node.Name] = 0
    nc.lastNodeUpdate[node.Name] = time.Now()
    return nil
}

func (nc *NodeCache) RemoveNode(nodeName string) error {
    nc.Lock()
    defer nc.Unlock()

    if _, exists := nc.nodes[nodeName]; !exists {
        return fmt.Errorf("node %s not found", nodeName)
    }

    delete(nc.nodes, nodeName)
    delete(nc.gpuAllocations, nodeName)
    delete(nc.lastNodeUpdate, nodeName)
    return nil
}

func (nc *NodeCache) UpdateGPUAllocation(nodeName string, gpuCount int) error {
    nc.Lock()
    defer nc.Unlock()

    if _, exists := nc.nodes[nodeName]; !exists {
        return fmt.Errorf("node %s not found", nodeName)
    }

    nc.gpuAllocations[nodeName] = gpuCount
    nc.lastNodeUpdate[nodeName] = time.Now()
    return nil
}

func (nc *NodeCache) GetNode(nodeName string) (*v1.Node, error) {
    nc.RLock()
    defer nc.RUnlock()

    node, exists := nc.nodes[nodeName]
    if !exists {
        return nil, fmt.Errorf("node %s not found", nodeName)
    }
    return node, nil
}

func (nc *NodeCache) GetGPUAllocation(nodeName string) (int, error) {
    nc.RLock()
    defer nc.RUnlock()

    if _, exists := nc.nodes[nodeName]; !exists {
        return 0, fmt.Errorf("node %s not found", nodeName)
    }
    return nc.gpuAllocations[nodeName], nil
}

func (nc *NodeCache) GetAllNodes() []*v1.Node {
    nc.RLock()
    defer nc.RUnlock()

    nodes := make([]*v1.Node, 0, len(nc.nodes))
    for _, node := range nc.nodes {
        nodes = append(nodes, node)
    }
    return nodes
}
