package kube

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
)

// GetKubernetesClient returns the client if its possible in cluster, otherwise tries to read HOME
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := GetKubernetesConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func GetKubernetesConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil && err == rest.ErrNotInCluster {
		// TODO: Read KUBECONFIG env variable as fallback
		config, err = clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), ".kube", "config"))
	}

	if err == nil {
		config.Burst = 1000
		config.QPS = 500
		config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(config.QPS, config.Burst)
		return config, nil
	}

	return nil, err
}
