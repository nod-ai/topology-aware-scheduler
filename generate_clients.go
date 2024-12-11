// +build generate

package main

import (
    "fmt"
    "os"
    "path/filepath"

    "k8s.io/code-generator/cmd/client-gen"
    clientgenargs "k8s.io/code-generator/cmd/client-gen/args"
    "k8s.io/code-generator/cmd/informer-gen"
    informergenargs "k8s.io/code-generator/cmd/informer-gen/args"
    "k8s.io/klog/v2"
)

func main() {
    // Paths
    apisPath := filepath.Join("pkg", "apis")
    clientsetPath := filepath.Join("pkg", "generated", "clientset")
    informersPath := filepath.Join("pkg", "generated", "informers")

    // Generate clientset
    clientGenArgs := &clientgenargs.GeneratorArgs{
        InputPackages:    []string{apisPath + "/topology/v1alpha1"},
        OutputPackage:    clientsetPath,
        ClientSetName:    "versioned",
        BoilerplateFile:  "boilerplate.go.txt",
    }
    if err := client_gen.Run(clientGenArgs); err != nil {
        klog.Fatalf("Failed to generate clientset: %v", err)
    }

    // Generate informers
    informerGenArgs := &informergenargs.GeneratorArgs{
        VersionedClientSetPackage: fmt.Sprintf("%s/versioned", clientsetPath),
        InternalClientSetPackage:  fmt.Sprintf("%s/internalclientset", clientsetPath),
        ListerPackage:             fmt.Sprintf("%s/listers", informersPath),
        InformerPackage:           fmt.Sprintf("%s/externalversions", informersPath),
        InputDirectories:          []string{apisPath + "/topology/v1alpha1"},
        OutputPackagePath:         informersPath,
        BoilerplateFile:           "boilerplate.go.txt",
    }
    if err := informer_gen.Run(informerGenArgs); err != nil {
        klog.Fatalf("Failed to generate informers: %v", err)
    }
}
