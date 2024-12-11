#!/usr/bin/env bash

# Copyright YEAR The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")

# Paths
APIS_PATH="${SCRIPT_ROOT}/pkg/apis"
CLIENTSET_PATH="${SCRIPT_ROOT}/pkg/generated/clientset"
INFORMERS_PATH="${SCRIPT_ROOT}/pkg/generated/informers"

# Find the location of the code-generator binaries
CLIENT_GEN=$(go env GOPATH)/bin/client-gen
INFORMER_GEN=$(go env GOPATH)/bin/informer-gen

# Generate clientset
$CLIENT_GEN \
  --clientset-name "versioned" \
  --input-base "${APIS_PATH}" \
  --input "topology/v1alpha1" \
  --output-package "${CLIENTSET_PATH}" \
  --go-header-file "${SCRIPT_ROOT}/boilerplate.go.txt"

# Generate informers
$INFORMER_GEN \
  --versioned-clientset-package "${CLIENTSET_PATH}/versioned" \
  --internal-clientset-package "${CLIENTSET_PATH}/internalclientset" \
  --listers-package "${INFORMERS_PATH}/listers" \
  --informers-package "${INFORMERS_PATH}/externalversions" \
  --input-dirs "${APIS_PATH}/topology/v1alpha1" \
  --output-package "${INFORMERS_PATH}" \
  --go-header-file "${SCRIPT_ROOT}/boilerplate.go.txt"
