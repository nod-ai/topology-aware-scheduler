# GPU Node Inference Setup

This guide explains how to set up inference workloads across multiple GPU nodes using JobSets in Kubernetes,
ensuring proper GPU isolation and resource management.

## Prerequisites

- Kubernetes cluster with multiple GPU nodes
- AMD GPUs configured with ROCm
- Local storage provisioner installed
- JobSet controller installed

## Quick Start

1. Label GPU nodes:
```bash
# Label each node with its specific GPU
for i in {0..7}; do
    kubectl label node <node-name-$i> gpu.amd.com/gpu-$i=true
done
```

2. Create PVCs:
```bash
# Create unique PVC for each GPU
for i in {0..7}; do
    sed "s/\${JOB_COMPLETION_INDEX}/$i/" pvc-template.yaml | kubectl apply -f -
done
```

3. Deploy JobSet:
```bash
kubectl apply -f jobset.yaml
```

## Configuration Features

- One-to-one mapping between pods and GPU nodes
- Unique PVC per GPU
- LoadBalancer for external access
- ROCm device isolation
- Local storage for model weights

## Verification

Check deployment status:
```bash
kubectl get jobset
kubectl get pods -l app=inference-service -o wide
kubectl get pvc | grep model-weights
```

## Notes

- Each pod gets dedicated GPU access through node affinity rules
- PVCs are created with ReadWriteOnce access mode
- Local storage ensures data locality
- Pod anti-affinity prevents double-scheduling on same node
