package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"strings"
)

func main() {
	fmt.Println("----- 7-indexer-informer -----")

	lw := newConfigMapsListerWatcher()
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	indexer, informer := cache.NewIndexerInformer(lw, &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("created: %s\n", configMap.Name)
		},
	}, indexers)

	stopper := make(chan struct{})
	defer close(stopper)

	fmt.Println("Start syncing....")

	go informer.Run(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		panic("timed out waiting for caches to sync")
	}

	keys, err := indexer.IndexKeys(cache.NamespaceIndex, "tmp")
	magicconch.Must(err)
	for _, k := range keys {
		fmt.Println(k)
	}

	startProbeServer(func(message string) string {
		keys, err := indexer.IndexKeys(cache.NamespaceIndex, "tmp")
		if err != nil {
			return "Error: " + err.Error()
		}
		return strings.Join(keys, ", ")
	})

	<-stopper
}
