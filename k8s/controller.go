package k8s

import (
	"context"
	"fmt"
	"github.com/invakid404/magicauth/oauth"
	"github.com/mitchellh/mapstructure"
	"github.com/ory/fosite"
	"github.com/stoewer/go-strcase"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func RunController(oauth *oauth.OAuth) (chan struct{}, error) {
	hasher := oauth.Provider.(*fosite.Fosite).Config.GetSecretsHasher(context.Background())

	config, err := clientcmd.BuildConfigFromFlags("", "")
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

	toFositeClient := func(obj *unstructured.Unstructured) (*fosite.DefaultClient, error) {
		spec := obj.UnstructuredContent()["spec"].(map[string]any)

		// Convert keys to snake case
		keys := make([]string, 0, len(spec))
		for key := range spec {
			keys = append(keys, key)
		}

		for _, key := range keys {
			newKey := strcase.SnakeCase(key)
			if key == newKey {
				continue
			}

			spec[newKey] = spec[key]
			delete(spec, key)
		}

		hashedSecret, err := hasher.Hash(context.Background(), []byte(spec["client_secret"].(string)))
		if err != nil {
			return nil, fmt.Errorf("failed to hash secret: %w", err)
		}

		spec["client_secret"] = hashedSecret

		var client fosite.DefaultClient
		config := &mapstructure.DecoderConfig{
			Metadata: nil,
			Result:   &client,
			TagName:  "json",
		}

		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create mapstructure decoder: %w", err)
		}

		if err := decoder.Decode(spec); err != nil {
			return nil, fmt.Errorf("failed to map oauth client: %w", err)
		}

		client.ID = obj.GetName()

		return &client, nil
	}

	upsertClient := func(obj *unstructured.Unstructured) {
		client, err := toFositeClient(obj)
		if err != nil {
			panic(err)
		}

		oauth.UpsertClient(client)
	}

	deleteClient := func(obj *unstructured.Unstructured) {
		oauth.DeleteClient(obj.GetName())
	}

	_, err = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			upsertClient(obj.(*unstructured.Unstructured))
		},
		UpdateFunc: func(oldObj, newObj any) {
			upsertClient(newObj.(*unstructured.Unstructured))
		},
		DeleteFunc: func(obj any) {
			deleteClient(obj.(*unstructured.Unstructured))
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
