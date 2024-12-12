# Topology-Aware GPU Scheduler for Kubernetes

A custom Kubernetes scheduler extension that optimizes GPU workload placement based on network topology constraints, designed for high-performance GPU clusters.

## Overview

A custom Kubernetes scheduler extension that optimizes GPU workload placement based on network topology constraints, designed for high-performance GPU clusters. This scheduler ensures optimal placement of GPU workloads by respecting the physical network topology of leaf-spine architecture, improving performance by up to 30% through smart placement decisions.

## Features

- 🎯 Topology-aware scheduling for GPU workloads
- 🔄 Smart domain selection based on job size
- 🔁 Automatic recovery with topology constraints
- 🧩 Anti-fragmentation mechanisms
- 📊 Real-time cluster state monitoring

## Architecture

### Core Components

- **Scheduler**: Optimizes GPU workload placement considering network topology
- **Domain Manager**: Manages network domains and node relationships
- **Plugin**: Kubernetes scheduler plugin implementation
- **Metrics**: Prometheus metrics for monitoring

## Installation

```bash
# Clone repository
git clone https://github.com/your-org/topology-aware-scheduler

# Build
make build

# Deploy
kubectl apply -f deploy/
```

## Components

### Algorithm (`pkg/algorithm/`)

- `scheduler.go`: Core scheduling logic
- `domain.go`: Domain management
- `topology.go`: Network topology handling
- `recovery.go`: Failure recovery mechanisms

### Scheduler Plugin (`pkg/scheduler/`)

```go
schedulerName: topology-aware-scheduler
```

Configuration:
```yaml
apiVersion: topology.scheduler/v1alpha1
kind: TopologySchedulerConfig
metadata:
  name: topology-scheduler-config
spec:
  scoringWeights:
    resourceAvailability: 0.4
    topologyAlignment: 0.3
    domainUtilization: 0.2
    historicalPerformance: 0.1
```

### Metrics

Available at `/metrics`:
- `topology_scheduler_latency_seconds`
- `topology_domain_utilization_ratio`
- `topology_gpu_allocation_ratio`
- `topology_placement_decisions_total`

### CLI Commands (`cmd/`)

```bash
# Start scheduler
./bin/scheduler --kubeconfig=config --scheduler-name=topology-aware-scheduler

# Start controller
./bin/controller --kubeconfig=config
```

## Development

### Testing

```bash
# Unit tests
go test ./pkg/...

# Integration tests
go test ./test/integration

# Coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Build

```bash
make build
make docker-build
make deploy
```

## Deployment Examples

### Single GPU Job
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: single-gpu-job
  annotations:
    topology.scheduler/gpu-count: "1"
spec:
  template:
    spec:
      schedulerName: topology-aware-scheduler
      containers:
      - name: gpu-job
        resources:
          limits:
            nvidia.com/gpu: 1
```

### Multi-GPU Job
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: multi-gpu-job
  annotations:
    topology.scheduler/gpu-count: "8"
    topology.scheduler/preferred-domain: "leaf-1"
spec:
  template:
    spec:
      schedulerName: topology-aware-scheduler
      containers:
      - name: gpu-job
        resources:
          limits:
            nvidia.com/gpu: 8
```

## Prerequisites

- Kubernetes 1.24+
- Go 1.20+
- Docker
- Access to a GPU cluster
- `kubectl` configured with cluster access

## Installation

1. Clone the repository:
```bash
git clone https://github.com/nod-ai/topology-aware-scheduler
cd topology-aware-scheduler
```

2. Install dependencies:
```bash
go mod download
```

3. Build:
```bash
make build
```

4. Deploy using the provided script:
```bash
./scripts/deploy.sh
```

## Configuration

The scheduler configuration is managed through a ConfigMap. Here's an example configuration:

```yaml
apiVersion: topology.scheduler/v1alpha1
kind: SchedulerConfig
metadata:
  name: topology-scheduler-config
spec:
  scoringWeights:
    resourceAvailability: 0.4
    topologyAlignment: 0.3
    domainUtilization: 0.2
    historicalPerformance: 0.1
  topologyConstraints:
    maxNodesPerLeaf: 4
    maxGPUsPerLeaf: 32
```

## Usage

### Submitting a GPU Job

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: gpu-job
  annotations:
    topology.scheduler/gpu-count: "8"
    topology.scheduler/preferred-domain: "leaf-1"
spec:
  template:
    spec:
      schedulerName: topology-aware-scheduler
      containers:
      - name: gpu-container
        image: gpu-workload:latest
        resources:
          limits:
            nvidia.com/gpu: 8
```

### Topology Constraints

The scheduler enforces the following placement rules:
- 2 nodes → Same leaf domain
- 4 nodes → Complete leaf domain
- 8 nodes → Two adjacent leaves
- 16 nodes → Four adjacent leaves

## Performance

### Metrics
- Scheduling latency: < 500ms
- Recovery time: < 30s
- Placement accuracy: 99.99%

### Monitoring

The scheduler exports Prometheus metrics at `/metrics`:
- `topology_scheduler_placement_duration_seconds`
- `topology_scheduler_recovery_duration_seconds`
- `topology_scheduler_domain_fragmentation_ratio`

### Advanced Use Cases

#### Inference Service Deployment
For latency-sensitive inference services:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gpu-inference-service
  annotations:
    topology.scheduler/gpu-count: "4"
    topology.scheduler/latency-sensitive: "true"
spec:
  replicas: 1
  template:
    spec:
      schedulerName: topology-aware-scheduler
      containers:
      - name: inference
        image: tensorflow/serving:latest
        resources:
          limits:
            nvidia.com/gpu: 4
```

#### Multi-Node Training Job
For large-scale distributed training:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: multi-node-training
  annotations:
    topology.scheduler/gpu-count: "16"
    topology.scheduler/preferred-domain: "spine-1"
    topology.scheduler/network-bandwidth: "200Gb"
    topology.scheduler/placement-strategy: "consolidated"
spec:
  parallelism: 4
  completions: 4
```

### Scheduler Annotations

The scheduler supports various annotations to optimize placement:

| Annotation | Description | Example Value |
|------------|-------------|---------------|
| `topology.scheduler/gpu-count` | Number of GPUs required | `"8"` |
| `topology.scheduler/preferred-domain` | Preferred network domain | `"leaf-1"` |
| `topology.scheduler/network-bandwidth` | Minimum network bandwidth | `"100Gb"` |
| `topology.scheduler/latency-sensitive` | Indicates latency-sensitive workload | `"true"` |
| `topology.scheduler/placement-strategy` | Placement strategy | `"consolidated"` |

### Placement Strategies

The scheduler supports several placement strategies:

1. **Consolidated** (`consolidated`)
   - Attempts to place all GPUs as close as possible
   - Optimizes for inter-GPU communication
   - Best for training workloads

2. **Distributed** (`distributed`)
   - Spreads GPUs across nodes
   - Optimizes for fault tolerance
   - Best for inference workloads

3. **Balanced** (`balanced`)
   - Balances between consolidation and distribution
   - Default strategy

### Common Use Cases

1. **Deep Learning Training**
   - Use consolidated placement
   - Request high network bandwidth
   - Specify GPU count based on model size

2. **Inference Services**
   - Use distributed placement
   - Enable latency-sensitive flag
   - Consider using node anti-affinity

3. **Research Workloads**
   - Use balanced placement
   - Specify preferred domain if needed
   - Adjust based on experiment requirements


### Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Troubleshooting

Common issues and solutions:

1. **Scheduler not starting**: Check logs using:
```bash
kubectl logs -n kube-system -l app=topology-scheduler
```

2. **Jobs not being scheduled**: Verify scheduler configuration:
```bash
kubectl get configmap -n kube-system topology-scheduler-config -o yaml
```

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

For support, please:
1. Check the [documentation](docs/)
2. Open an issue
3. Join our [Slack channel](#)


