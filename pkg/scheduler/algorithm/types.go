package algorithm

import (
    v1 "k8s.io/api/core/v1"
)

// TopologyScore represents the scoring weights for different factors
type TopologyScore struct {
    ResourceAvailability float64
    TopologyAlignment    float64
    DomainUtilization   float64
    HistoricalPerf      float64
}

// Domain represents a leaf switch domain containing nodes
type Domain struct {
    ID          string
    Nodes       []*v1.Node
    TotalGPUs   int
    UsedGPUs    int
    LeafSwitch  string
    SpineSwitch string
}

// TopologyState represents the current state of the cluster topology
type TopologyState struct {
    Domains          map[string]*Domain
    SpineConnections map[string][]string
}