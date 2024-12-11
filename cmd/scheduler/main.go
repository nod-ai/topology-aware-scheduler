// +build !generate
package main

import (
    "context"
    "flag"
    "fmt"
    "net/http"
    "os"
    "time"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/tools/leaderelection"
    "k8s.io/client-go/tools/leaderelection/resourcelock"
    "k8s.io/klog/v2"
    "k8s.io/kubernetes/pkg/scheduler/apis/config"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "github.com/prometheus/client_golang/prometheus/promhttp"

    "github.com/yourusername/topology-aware-gpu-scheduler/pkg/scheduler/algorithm"
    clientset "github.com/yourusername/topology-aware-gpu-scheduler/pkg/generated/clientset/versioned"
)

var (
    masterURL            string
    kubeconfig          string
    schedulerName       string
    leaderElect         bool
    lockObjectName      string
    lockObjectNamespace string
    version            string // Added for version info
    buildDate          string // Added for build date
)

func main() {
    klog.InitFlags(nil)
    flag.Parse()

    klog.Infof("Starting Topology-Aware GPU Scheduler - Version: %s, Build Date: %s", version, buildDate)

    // Build kubernetes config
    cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
    if err != nil {
        klog.Fatalf("Error building kubeconfig: %v", err)
    }

    // Create kubernetes clientset
    kubeClient, err := kubernetes.NewForConfig(cfg)
    if err != nil {
        klog.Fatalf("Error building kubernetes clientset: %v", err)
    }

    // Create topology clientset
    topologyClient, err := clientset.NewForConfig(cfg)
    if err != nil {
        klog.Fatalf("Error building topology clientset: %v", err)
    }

    // Create scheduler cache and topology cache
    nodeCache := algorithm.NewNodeCache()
    topologyCache := algorithm.NewTopologyCache(nodeCache)
    
    // Create the scheduler
    scheduler := algorithm.NewTopologyScheduler(topologyCache)

    // Start metrics server
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        klog.Fatal(http.ListenAndServe(":8080", nil))
    }()

    // Leader election handling
    if leaderElect {
        lock := &resourcelock.LeaseLock{
            LeaseMeta: metav1.ObjectMeta{
                Name:      lockObjectName,
                Namespace: lockObjectNamespace,
            },
            Client: kubeClient.CoordinationV1(),
            LockConfig: resourcelock.ResourceLockConfig{
                Identity: getHostname(),
            },
        }

        leaderelection.RunOrDie(context.Background(), leaderelection.LeaderElectionConfig{
            Lock:            lock,
            LeaseDuration:  15 * time.Second,
            RenewDeadline:  10 * time.Second,
            RetryPeriod:    2 * time.Second,
            Callbacks: leaderelection.LeaderCallbacks{
                OnStartedLeading: func(ctx context.Context) {
                    runScheduler(scheduler, kubeClient)
                },
                OnStoppedLeading: func() {
                    klog.Info("Leader lost")
                    os.Exit(0)
                },
                OnNewLeader: func(identity string) {
                    klog.Infof("New leader elected: %v", identity)
                },
            },
        })
    } else {
        runScheduler(scheduler, kubeClient)
    }
}

func runScheduler(scheduler *algorithm.TopologyScheduler, client kubernetes.Interface) {
    sched := scheduler.NewScheduler(client)
    
    // Start scheduling loop
    stopCh := make(chan struct{})
    defer close(stopCh)
    
    go sched.Run(stopCh)
    
    // Start node monitor
    monitor := sched.GetMonitor()
    go monitor.Start()
    
    <-stopCh
}

func getHostname() string {
    hostname, err := os.Hostname()
    if err != nil {
        return "unknown"
    }
    return hostname
}

func init() {
    flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
    flag.StringVar(&masterURL, "master", "", "Kubernetes API server address")
    flag.StringVar(&schedulerName, "scheduler-name", "topology-aware-scheduler", "Name of the scheduler")
    flag.BoolVar(&leaderElect, "leader-elect", true, "Enable leader election")
    flag.StringVar(&lockObjectName, "lock-object-name", "topology-scheduler", "Name of lock object")
    flag.StringVar(&lockObjectNamespace, "lock-object-namespace", "kube-system", "Namespace of lock object")
}
