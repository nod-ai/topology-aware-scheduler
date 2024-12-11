#!/bin/bash
# setup-codegen.sh

# Create necessary directories
mkdir -p hack
mkdir -p pkg/apis/topology/v1alpha1
mkdir -p pkg/generated

# Clone code-generator temporarily
git clone https://github.com/kubernetes/code-generator.git /tmp/code-generator
cp /tmp/code-generator/generate-groups.sh hack/generate-groups.sh
chmod +x hack/generate-groups.sh

# Create boilerplate header file
cat > hack/boilerplate.go.txt << EOL
/*
Copyright $(date +%Y) The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
EOL

# Create the update-codegen.sh script
cat > hack/update-codegen.sh << 'EOL'
#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

bash "${SCRIPT_ROOT}"/hack/generate-groups.sh "deepcopy,client,informer,lister" \
  github.com/yourusername/topology-aware-gpu-scheduler/pkg/generated \
  github.com/yourusername/topology-aware-gpu-scheduler/pkg/apis \
  topology:v1alpha1 \
  --output-base "$(dirname "${BASH_SOURCE[0]}")/../../../.." \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt
EOL

chmod +x hack/update-codegen.sh

# Clean up temporary clone
rm -rf /tmp/code-generator
