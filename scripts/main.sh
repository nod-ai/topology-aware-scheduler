#!/bin/bash
set -e

# Configuration
REGISTRY=${REGISTRY:-"localhost:5000"}
IMAGE_NAME=${IMAGE_NAME:-"topology-scheduler"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}
NAMESPACE=${NAMESPACE:-"kube-system"}

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Function to apply Kubernetes resources
apply_resources() {
    local dir=$1
    local resource_type=$2
    
    echo -e "${GREEN}Applying $resource_type...${NC}"
    for file in $dir/*.yaml; do
        if [ -f "$file" ]; then
            echo "Applying $file"
            kubectl apply -f "$file"
        fi
    done
}

echo "Building and deploying Topology-Aware GPU Scheduler..."

# Build Docker image
echo -e "${GREEN}Building Docker image...${NC}"
docker build -t ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG} .

# Push to registry
echo -e "${GREEN}Pushing image to registry...${NC}"
docker push ${REGISTRY}/${IMAGE_NAME}:${IMAGE_TAG}

# Create namespace if it doesn't exist
kubectl get namespace ${NAMESPACE} || kubectl create namespace ${NAMESPACE}

# Apply resources in order
apply_resources "deploy/crds" "Custom Resource Definitions"
apply_resources "deploy/rbac" "RBAC configuration"
apply_resources "deploy/config" "ConfigMaps"
apply_resources "deploy/scheduler" "Scheduler components"

# Wait for deployment to be ready
echo -e "${GREEN}Waiting for deployment to be ready...${NC}"
kubectl rollout status deployment/topology-scheduler -n ${NAMESPACE}

# Verify deployment
echo -e "${GREEN}Verifying deployment...${NC}"
if kubectl get pods -n ${NAMESPACE} -l app=topology-scheduler | grep -q Running; then
    echo -e "${GREEN}Deployment successful!${NC}"
    echo "You can now submit GPU jobs using the topology-aware-scheduler"
    echo "Example:"
    echo "kubectl apply -f examples/gpu-job.yaml"
else
    echo -e "${RED}Deployment failed. Please check the logs:${NC}"
    echo "kubectl logs -n ${NAMESPACE} -l app=topology-scheduler"
    exit 1
fi

# Print metrics endpoint
echo -e "\n${GREEN}Metrics are available at:${NC}"
echo "http://topology-scheduler-metrics.${NAMESPACE}:8080/metrics"

# Installation verification
echo -e "\n${GREEN}Running verification tests...${NC}"
./scripts/verify-installation.sh

echo -e "\n${GREEN}Setup complete!${NC}"
