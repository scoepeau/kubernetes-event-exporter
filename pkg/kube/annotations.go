package kube

import (
	"strings"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type AnnotationRetriever interface {
	GetAnnotations(*v1.ObjectReference) (map[string]string, error)
}

type AnnotationCache struct {
	dynClient dynamic.Interface
	clientset *kubernetes.Clientset

	cache *lru.ARCCache
	sync.RWMutex
}

func NewAnnotationCache(kubeconfig *rest.Config) AnnotationRetriever {
	cache, err := lru.NewARC(1024)
	if err != nil {
		panic("cannot init cache: " + err.Error())
	}
	return &AnnotationCache{
		dynClient: dynamic.NewForConfigOrDie(kubeconfig),
		clientset: kubernetes.NewForConfigOrDie(kubeconfig),
		cache:     cache,
	}
}

func (a *AnnotationCache) GetAnnotations(reference *v1.ObjectReference) (map[string]string, error) {
	uid := reference.UID

	if val, ok := a.cache.Get(uid); ok {
		return val.(map[string]string), nil
	}

	obj, err := GetObject(reference, a.clientset, a.dynClient)
	if err == nil {
		annotations := obj.GetAnnotations()
		for key := range annotations {
			if strings.Contains(key, "kubernetes.io/") || strings.Contains(key, "k8s.io/") {
				delete(annotations, key)
			}
		}
		a.cache.Add(uid, annotations)
		return annotations, nil
	}

	if errors.IsNotFound(err) {
		var empty map[string]string
		a.cache.Add(uid, empty)
		return nil, nil
	}

	return nil, err

}

type AnnotationWithoutCache struct {
	dynClient dynamic.Interface
	clientset *kubernetes.Clientset
}

func NewAnnotationWithoutCache(kubeconfig *rest.Config) AnnotationRetriever {
	return &AnnotationWithoutCache{
		dynClient: dynamic.NewForConfigOrDie(kubeconfig),
		clientset: kubernetes.NewForConfigOrDie(kubeconfig),
	}
}

func (a *AnnotationWithoutCache) GetAnnotations(reference *v1.ObjectReference) (map[string]string, error) {
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
