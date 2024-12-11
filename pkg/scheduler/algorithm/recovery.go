package algorithm

import (
    "fmt"
    "sync"
    "context"
    v1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RecoveryManager struct {
    domainManager *DomainManager
    scheduler     *TopologyScheduler
    recoveryLock  sync.Mutex
}

func NewRecoveryManager(dm *DomainManager, scheduler *TopologyScheduler) *RecoveryManager {
    return &RecoveryManager{
        domainManager: dm,
        scheduler:     scheduler,
    }
}

func (rm *RecoveryManager) HandleNodeFailure(node *v1.Node, pods []*v1.Pod) error {
    rm.recoveryLock.Lock()
    defer rm.recoveryLock.Unlock()

    // Get the failed node's domain
    domain, err := rm.domainManager.GetDomainByNode(node.Name)
    if err != nil {
        return fmt.Errorf("failed to get domain for node %s: %v", node.Name, err)
    }

    // Group pods by GPU requirements
    gpuPods, nonGpuPods := rm.categorizePods(pods)

    // Handle GPU pods first as they are more critical
    if err := rm.recoverGPUPods(domain, gpuPods); err != nil {
        return fmt.Errorf("failed to recover GPU pods: %v", err)
    }

    // Handle non-GPU pods
    if err := rm.recoverNonGPUPods(domain, nonGpuPods); err != nil {
        return fmt.Errorf("failed to recover non-GPU pods: %v", err)
    }

    // Update domain state
    return rm.domainManager.HandleNodeRemoval(node.Name)
}

func (rm *RecoveryManager) categorizePods(pods []*v1.Pod) (gpuPods []*v1.Pod, nonGpuPods []*v1.Pod) {
    for _, pod := range pods {
        if requiresGPU(pod) {
            gpuPods = append(gpuPods, pod)
        } else {
            nonGpuPods = append(nonGpuPods, pod)
        }
    }
    return
}

func (rm *RecoveryManager) recoverGPUPods(failedDomain *Domain, pods []*v1.Pod) error {
    // Sort pods by priority and GPU requirements
    sortPodsByPriority(pods)

    for _, pod := range pods {
        // Get GPU requirements
        gpuCount := getGPURequirements(pod)
        
        // Try to schedule in the same domain first
        newNode, err := rm.scheduler.FindNodeInDomain(failedDomain, gpuCount)
        if err == nil {
            if err := rm.migratePod(pod, newNode); err != nil {
                return err
            }
            continue
        }

        // Try adjacent domains if same domain failed
        adjacentDomains := rm.domainManager.GetAdjacentDomains(failedDomain.Name)
        scheduled := false
        for _, adjDomain := range adjacentDomains {
            newNode, err := rm.scheduler.FindNodeInDomain(adjDomain, gpuCount)
            if err == nil {
                if err := rm.migratePod(pod, newNode); err != nil {
                    return err
                }
                scheduled = true
                break
            }
        }

        if !scheduled {
            // Fall back to any available domain
            newNode, err := rm.scheduler.FindNodeForGPUs(gpuCount)
            if err != nil {
                return fmt.Errorf("failed to find replacement node for pod %s: %v", pod.Name, err)
            }
            if err := rm.migratePod(pod, newNode); err != nil {
                return err
            }
        }
    }
    return nil
}

func (rm *RecoveryManager) recoverNonGPUPods(failedDomain *Domain, pods []*v1.Pod) error {
    sortPodsByPriority(pods)

    for _, pod := range pods {
        newNode, err := rm.scheduler.FindNodeForPod(pod)
        if err != nil {
            return fmt.Errorf("failed to find replacement node for pod %s: %v", pod.Name, err)
        }
        
        if err := rm.migratePod(pod, newNode); err != nil {
            return err
        }
    }
    return nil
}

func (rm *RecoveryManager) migratePod(pod *v1.Pod, newNode *v1.Node) error {
    // Create new pod spec with updated node name
    newPod := pod.DeepCopy()
    newPod.Spec.NodeName = newNode.Name
    newPod.ResourceVersion = ""
    newPod.UID = ""
    newPod.Status = v1.PodStatus{}

    // Set scheduling timestamp
    now := metav1.Now()
    newPod.Status.StartTime = &now

    // Delete old pod and create new one
    err := rm.scheduler.DeletePod(context.Background(), pod)
    if err != nil {
        return fmt.Errorf("failed to delete old pod %s: %v", pod.Name, err)
    }

    _, err = rm.scheduler.CreatePod(context.Background(), newPod)
    if err != nil {
        return fmt.Errorf("failed to create new pod %s on node %s: %v", 
            newPod.Name, newNode.Name, err)
    }

    return nil
}

func requiresGPU(pod *v1.Pod) bool {
    for _, container := range pod.Spec.Containers {
        if _, ok := container.Resources.Limits["nvidia.com/gpu"]; ok {
            return true
        }
    }
    return false
}

func getGPURequirements(pod *v1.Pod) int {
    var total int
    for _, container := range pod.Spec.Containers {
        if gpus, ok := container.Resources.Limits["nvidia.com/gpu"]; ok {
            total += int(gpus.Value())
        }
    }
    return total
}

func sortPodsByPriority(pods []*v1.Pod) {
    sort.Slice(pods, func(i, j int) bool {
        iPriority := pods[i].Spec.Priority
        jPriority := pods[j].Spec.Priority
        
        if iPriority == nil && jPriority == nil {
            return false
        }
        if iPriority == nil {
            return false
        }
        if jPriority == nil {
            return true
        }
        
        return *iPriority > *jPriority
    })
}
