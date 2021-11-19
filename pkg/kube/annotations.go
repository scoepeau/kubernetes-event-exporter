package kube

import (
	lru "github.com/hashicorp/golang-lru"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
	"sync"
)

type AnnotationCache struct {
	dynClient dynamic.Interface
	clientset *kubernetes.Clientset

	sync.RWMutex
}

func NewAnnotationCache(kubeconfig *rest.Config) *AnnotationCache {
	return &AnnotationCache{
		dynClient: dynamic.NewForConfigOrDie(kubeconfig),
		clientset: kubernetes.NewForConfigOrDie(kubeconfig),
	}
}

func (a *AnnotationCache) GetAnnotationsWithCache(reference *v1.ObjectReference) (map[string]string, error) {
	uid := reference.UID

	obj, err := GetObject(reference, a.clientset, a.dynClient)
	if err == nil {
		annotations := obj.GetAnnotations()
		for key := range annotations {
			if strings.Contains(key, "kubernetes.io/") || strings.Contains(key, "k8s.io/") {
				delete(annotations, key)
			}
		}
		return annotations, nil
	}

	if errors.IsNotFound(err) {
		return nil, nil
	}

	return nil, err

}
