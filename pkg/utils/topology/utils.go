go
package topology

import (
    "encoding/json"
    "fmt"
    "math"
    "strconv"
    v1 "k8s.io/api/core/v1"
)

type NodeGPUInfo struct {
    TotalGPUs     int
    AllocatedGPUs int
    GPUTypes      []string
    GPUMemory     []int64
}

func ExtractNodeGPUInfo(node *v1.Node) (*NodeGPUInfo, error) {
    info := &NodeGPUInfo{}
    
    if val, ok := node.Labels["nvidia.com/gpu.count"]; ok {
        count, err := strconv.Atoi(val)
        if err != nil {
            return nil, fmt.Errorf("invalid GPU count: %v", err)
        }
        info.TotalGPUs = count
    }

    if val, ok := node.Labels["nvidia.com/gpu.memory"]; ok {
        if err := json.Unmarshal([]byte(val), &info.GPUMemory); err != nil {
            return nil, fmt.Errorf("invalid GPU memory info: %v", err)
        }
    }
    
    if val, ok := node.Labels["nvidia.com/gpu.type"]; ok {
        info.GPUTypes = append(info.GPUTypes, val)
    }

    return info, nil
}

func CalculateDomainDistance(source, target *Domain, connections map[string][]string) int {
    if source.Name == target.Name {
        return 0
    }

    visited := make(map[string]bool)
    queue := []struct {
        domain string
        dist   int
    }{{source.Name, 0}}
    visited[source.Name] = true

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        for _, neighbor := range connections[current.domain] {
            if neighbor == target.Name {
                return current.dist + 1
            }
            if !visited[neighbor] {
                visited[neighbor] = true
                queue = append(queue, struct {
                    domain string
                    dist   int
                }{neighbor, current.dist + 1})
            }
        }
    }
    return math.MaxInt32
}

func CalculateDomainHealth(domain *Domain) float64 {
    if len(domain.Nodes) == 0 {
        return 0.0
    }

    healthyNodes := 0
    for _, node := range domain.Nodes {
        if IsNodeHealthy(node) {
            healthyNodes++
        }
    }
    return float64(healthyNodes) / float64(len(domain.Nodes))
}

func IsNodeHealthy(node *v1.Node) bool {
    for _, condition := range node.Status.Conditions {
        if condition.Type == v1.NodeReady {
            return condition.Status == v1.ConditionTrue
        }
    }
    return false
}

func CalculateGPUFragmentation(domain *Domain) float64 {
    if len(domain.Nodes) == 0 {
        return 0.0
    }

    totalPartialNodes := 0
    for _, node := range domain.Nodes {
        info, err := ExtractNodeGPUInfo(node)
        if err != nil {
            continue
        }
        if info.AllocatedGPUs > 0 && info.AllocatedGPUs < info.TotalGPUs {
            totalPartialNodes++
        }
    }

    return float64(totalPartialNodes) / float64(len(domain.Nodes))
}
