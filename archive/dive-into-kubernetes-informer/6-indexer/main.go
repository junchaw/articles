package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func main() {
	lw := newConfigMapsListerWatcher()
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	indexer, informer := cache.NewIndexerInformer(
		lw, &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{}, indexers)

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
}
