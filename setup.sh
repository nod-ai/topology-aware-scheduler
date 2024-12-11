#!/bin/bash

# Create directory structure
mkdir -p pkg/apis/topology/v1alpha1
mkdir -p pkg/generated

# Create boilerplate header
cat > boilerplate.go.txt << 'EOF'
/*
Copyright 2024 The Kubernetes Authors.

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
EOF

# Create doc.go
cat > pkg/apis/topology/v1alpha1/doc.go << 'EOF'
// +k8s:deepcopy-gen=package
// +groupName=topology.scheduler.k8s.io

package v1alpha1
EOF

# Create types.go
cat > pkg/apis/topology/v1alpha1/types.go << 'EOF'
package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TopologyScheduler is a specification for a TopologyScheduler resource
type TopologyScheduler struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec   TopologySchedulerSpec   `json:"spec"`
    Status TopologySchedulerStatus `json:"status"`
}

// TopologySchedulerSpec defines the desired state of TopologyScheduler
type TopologySchedulerSpec struct {
    // Add your custom fields here
    GPUCount    int32  `json:"gpuCount"`
    Domain      string `json:"domain"`
    Utilization int32  `json:"utilization"`
}

// TopologySchedulerStatus defines the observed state of TopologyScheduler
type TopologySchedulerStatus struct {
    Phase      string `json:"phase"`
    Message    string `json:"message"`
    LastUpdate string `json:"lastUpdate"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TopologySchedulerList contains a list of TopologyScheduler
type TopologySchedulerList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items []TopologyScheduler `json:"items"`
}
EOF

# Create register.go
cat > pkg/apis/topology/v1alpha1/register.go << 'EOF'
package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/runtime/schema"
)

const (
    GroupName = "topology.scheduler.k8s.io"
    Version   = "v1alpha1"
)

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

func Resource(resource string) schema.GroupResource {
    return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
    SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
    AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
    scheme.AddKnownTypes(SchemeGroupVersion,
        &TopologyScheduler{},
        &TopologySchedulerList{},
    )
    metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
    return nil
}
EOF

# Create updated codegen.sh
cat > codegen.sh << 'EOF'
#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo $(go env GOPATH)/pkg/mod/k8s.io/code-generator@v0.28.0)}

# Create temporary directory for code generation
TMP_DIR=$(mktemp -d)
chmod 755 $TMP_DIR

echo "Generating client codes..."
bash "${CODEGEN_PKG}/kube_codegen.sh" \
  "client,lister,informer" \
  github.com/iamakanshab/topology-aware-gpu-scheduler/pkg/generated \
  github.com/iamakanshab/topology-aware-gpu-scheduler/pkg/apis \
  "topology:v1alpha1" \
  --output-base "${TMP_DIR}" \
  --go-header-file "${SCRIPT_ROOT}/boilerplate.go.txt"

# Ensure target directory exists
mkdir -p "${SCRIPT_ROOT}/pkg/generated"

# Copy generated files to the right location
cp -r "${TMP_DIR}/github.com/iamakanshab/topology-aware-gpu-scheduler/pkg/generated"/* "${SCRIPT_ROOT}/pkg/generated/"

# Cleanup
rm -rf "${TMP_DIR}"
EOF
