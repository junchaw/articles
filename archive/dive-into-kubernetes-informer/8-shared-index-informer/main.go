package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"strings"
)

func main() {
	fmt.Println("----- 8-shared-indexer-informer -----")

	lw := newConfigMapsListerWatcher()
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	informer := cache.NewSharedIndexInformer(lw, &corev1.ConfigMap{}, 0, indexers)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("created: %s\n", configMap.Name)
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		panic("timed out waiting for caches to sync")
	}

	keys, err := informer.GetIndexer().IndexKeys(cache.NamespaceIndex, "tmp")
	magicconch.Must(err)
	for _, k := range keys {
		fmt.Println(k)
	}

	startProbeServer(func(message string) string {
		keys, err := informer.GetIndexer().IndexKeys(cache.NamespaceIndex, "tmp")
		if err != nil {
			return "Error: " + err.Error()
		}
		return strings.Join(keys, ", ")
	})

	<-stopCh
}
