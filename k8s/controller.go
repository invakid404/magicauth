package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

func RunController() (chan struct{}, error) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/inva/.kube/config")
	if err != nil {
		return nil, fmt.Errorf("failed to build config: %w", err)
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	oauthClientSchema := schema.GroupVersionResource{
		Group:    "magicauth.invak.id",
		Version:  "v1",
		Resource: "oauthclients",
	}

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return clientset.Resource(oauthClientSchema).Namespace("").List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return clientset.Resource(oauthClientSchema).Namespace("").Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		0,
		cache.Indexers{},
	)

	_, err = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			log.Printf("resource created: %v", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			log.Printf("resource: %v was update to: %v", oldObj, newObj)
		},
		DeleteFunc: func(obj interface{}) {
			log.Printf("resource deleted: %v", obj)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add event handler: %w", err)
	}

	stop := make(chan struct{})
	go informer.Run(stop)

	if !cache.WaitForCacheSync(stop, informer.HasSynced) {
		return nil, fmt.Errorf("timeout waiting for cache sync")
	}

	return stop, nil
}
