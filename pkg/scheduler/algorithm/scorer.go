package scheduler

import (
    "k8s.io/api/core/v1"
    "math"
)

type Scorer struct {
    topology *TopologyManager
    weights  *ScoringWeights
}

type ScoringWeights struct {
    GPUUtilization     float64
    NetworkProximity   float64
    DomainAffinity     float64
    LoadBalance        float64
}

func NewScorer(topology *TopologyManager, weights *ScoringWeights) *Scorer {
    return &Scorer{
        topology: topology,
        weights:  weights,
    }
}

func (s *Scorer) ScoreNode(
    node *v1.Node,
    requirements *ResourceRequirements,
    constraints *SchedulingConstraints,
) float64 {
    gpuScore := s.scoreGPUUtilization(node, requirements)
    networkScore := s.scoreNetworkProximity(node, constraints)
    affinityScore := s.scoreDomainAffinity(node, constraints)
    loadScore := s.scoreLoadBalance(node)

    return (gpuScore * s.weights.GPUUtilization) +
           (networkScore * s.weights.NetworkProximity) +
           (affinityScore * s.weights.DomainAffinity) +
           (loadScore * s.weights.LoadBalance)
}
