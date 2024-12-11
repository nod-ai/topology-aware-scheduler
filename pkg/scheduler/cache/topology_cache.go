package algorithm

import (
    "fmt"
    "sync"
    "time"
    v1 "k8s.io/api/core/v1"
)

type TopologyCache struct {
    sync.RWMutex
    nodeCache         *NodeCache
    domains           map[string]*Domain
    spineConnections map[string][]string
    domainForNode    map[string]string
    lastUpdated      time.Time
}

func NewTopologyCache(nodeCache *NodeCache) *TopologyCache {
    return &TopologyCache{
        nodeCache:        nodeCache,
        domains:         make(map[string]*Domain),
        spineConnections: make(map[string][]string),
        domainForNode:   make(map[string]string),
        lastUpdated:     time.Now(),
    }
}

func (tc *TopologyCache) AddDomain(domain *Domain) error {
    tc.Lock()
    defer tc.Unlock()

    if _, exists := tc.domains[domain.Name]; exists {
        return fmt.Errorf("domain %s already exists", domain.Name)
    }

    tc.domains[domain.Name] = domain
    for _, node := range domain.Nodes {
        tc.domainForNode[node.Name] = domain.Name
    }
    tc.lastUpdated = time.Now()
    return nil
}

func (tc *TopologyCache) AddNodeToDomain(nodeName, domainName string) error {
    tc.Lock()
    defer tc.Unlock()

    domain, exists := tc.domains[domainName]
    if !exists {
        return fmt.Errorf("domain %s not found", domainName)
    }

    node, err := tc.nodeCache.GetNode(nodeName)
    if err != nil {
        return fmt.Errorf("node %s not found in node cache", nodeName)
    }

    domain.Nodes = append(domain.Nodes, node)
    tc.domainForNode[nodeName] = domainName
    tc.lastUpdated = time.Now()
    return nil
}

func (tc *TopologyCache) RemoveNodeFromDomain(nodeName, domainName string) error {
    tc.Lock()
    defer tc.Unlock()

    domain, exists := tc.domains[domainName]
    if !exists {
        return fmt.Errorf("domain %s not found", domainName)
    }

    for i, node := range domain.Nodes {
        if node.Name == nodeName {
            domain.Nodes = append(domain.Nodes[:i], domain.Nodes[i+1:]...)
            delete(tc.domainForNode, nodeName)
            tc.lastUpdated = time.Now()
            return nil
        }
    }
    return fmt.Errorf("node %s not found in domain %s", nodeName, domainName)
}

func (tc *TopologyCache) AddSpineConnection(source, target string) error {
    tc.Lock()
    defer tc.Unlock()

    if _, exists := tc.domains[source]; !exists {
        return fmt.Errorf("source domain %s not found", source)
    }
    if _, exists := tc.domains[target]; !exists {
        return fmt.Errorf("target domain %s not found", target)
    }

    tc.spineConnections[source] = append(tc.spineConnections[source], target)
    tc.lastUpdated = time.Now()
    return nil
}

func (tc *TopologyCache) GetDomainForNode(nodeName string) (*Domain, error) {
    tc.RLock()
    defer tc.RUnlock()

    domainName, exists := tc.domainForNode[nodeName]
    if !exists {
        return nil, fmt.Errorf("no domain found for node %s", nodeName)
    }
    return tc.domains[domainName], nil
}

func (tc *TopologyCache) GetConnectedDomains(domainName string) ([]*Domain, error) {
    tc.RLock()
    defer tc.RUnlock()

    connections, exists := tc.spineConnections[domainName]
    if !exists {
        return nil, fmt.Errorf("domain %s not found", domainName)
    }

    var connectedDomains []*Domain
    for _, conn := range connections {
        if domain, exists := tc.domains[conn]; exists {
            connectedDomains = append(connectedDomains, domain)
        }
    }
    return connectedDomains, nil
}

func (tc *TopologyCache) GetAllDomains() []*Domain {
    tc.RLock()
    defer tc.RUnlock()

    domains := make([]*Domain, 0, len(tc.domains))
    for _, domain := range tc.domains {
        domains = append(domains, domain)
    }
    return domains
}
