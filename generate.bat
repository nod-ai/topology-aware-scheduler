client-gen ^
  --clientset-name "versioned" ^
  --input-base "github.com/nod-ai/topology-aware-scheduler/pkg/apis" ^
  --input "topology/v1alpha1" ^
  --output-package "github.com/nod-ai/topology-aware-scheduler/pkg/generated/clientset" ^
  --go-header-file "github.com/nod-ai/topology-aware-scheduler/githuboilerplate.go.txt" ^
