go
// +build !generate
package main

import (
    "flag"
    "os"
    "time"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/klog/v2"
    "github.com/nod-ai/topology-aware-scheduler/pkg/apis/topology/v1alpha1"
    clientset "github.com/nod-ai/topology-aware-scheduler/pkg/generated/clientset/versioned"
    informers "github.com/nod-ai/topology-aware-scheduler/pkg/generated/informers/externalversions"
    listers "github.com/nod-ai/topology-aware-scheduler/pkg/generated/listers/topology/v1alpha1"
    "github.com/nod-ai/topology-aware-scheduler/pkg/controller"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "net/http"
)

var (
    masterURL  string
    kubeconfig string
)

func main() {
    klog.InitFlags(nil)
    flag.Parse()

    cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
    if err != nil {
        klog.Fatalf("Error building kubeconfig: %s", err.Error())
    }

    kubeClient, err := kubernetes.NewForConfig(cfg)
    if err != nil {
        klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
    }

    topologyClient, err := clientset.NewForConfig(cfg)
    if err != nil {
        klog.Fatalf("Error building topology clientset: %s", err.Error())
    }

    topologyInformerFactory := informers.NewSharedInformerFactory(topologyClient, time.Second*30)

    controller := controller.NewController(
        kubeClient,
        topologyClient,
        topologyInformerFactory.Topology().V1alpha1().TopologySchedulerConfigs(),
        topologyInformerFactory.Topology().V1alpha1().DomainConfigs(),
    )

    // Start metrics server
    go func() {
        http.Handle("/metrics", promhttp.Handler())
        klog.Fatal(http.ListenAndServe(":8080", nil))
    }()

    // Notice that there is no need to run Start methods in a separate goroutine.
    // Start() is non-blocking and runs the informer collection in the background.
    topologyInformerFactory.Start(stopCh)

    if err = controller.Run(2, stopCh); err != nil {
        klog.Fatalf("Error running controller: %s", err.Error())
    }
}

func init() {
    flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
    flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
