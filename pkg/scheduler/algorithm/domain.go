package scheduler

import (
    "fmt"
    "sync"
)

type DomainManager struct {
    mu      sync.RWMutex
    domains map[string]*Domain
}

type Domain struct {
    Name        string
    Type        string // "leaf" or "spine"
    Bandwidth   int64
    Latency     float64
    Nodes       map[string]*Node
    Parent      string
    Children    []string
}

func NewDomainManager() *DomainManager {
    return &DomainManager{
        domains: make(map[string]*Domain),
    }
}

func (dm *DomainManager) AddDomain(domain *Domain) error {
    dm.mu.Lock()
    defer dm.mu.Unlock()

    if _, exists := dm.domains[domain.Name]; exists {
        return fmt.Errorf("domain %s already exists", domain.Name)
    }

    dm.domains[domain.Name] = domain
    return nil
}

func (dm *DomainManager) GetDomainByNode(nodeName string) (*Domain, error) {
    dm.mu.RLock()
    defer dm.mu.RUnlock()

    for _, domain := range dm.domains {
        if _, exists := domain.Nodes[nodeName]; exists {
            return domain, nil
        }
    }
    return nil, fmt.Errorf("no domain found for node %s", nodeName)
}
