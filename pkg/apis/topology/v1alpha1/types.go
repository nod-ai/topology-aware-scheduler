package v1alpha1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:defaulter-gen=true
// +kubebuilder:resource:scope=Namespaced

// TopologyScheduler is a specification for a TopologyScheduler resource
type TopologyScheduler struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   TopologySchedulerSpec   `json:"spec"`
    Status TopologySchedulerStatus `json:"status"`
}

// TopologySchedulerSpec is the spec for a TopologyScheduler resource
type TopologySchedulerSpec struct {
    // Add your scheduler-specific fields here
    GPUCount    int32  `json:"gpuCount"`
    Domain      string `json:"domain"`
    Utilization int32  `json:"utilization"`
}

// TopologySchedulerStatus is the status for a TopologyScheduler resource
type TopologySchedulerStatus struct {
    // Add status fields here
    Phase      string `json:"phase"`
    Message    string `json:"message"`
    LastUpdate string `json:"lastUpdate"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TopologySchedulerList is a list of TopologyScheduler resources
type TopologySchedulerList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata"`
    Items []TopologyScheduler `json:"items"`
}
