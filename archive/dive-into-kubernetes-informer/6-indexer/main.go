package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func main() {
	fmt.Println("----- 6-indexer -----")

	lw := newConfigMapsListerWatcher()
	indexers := cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}
	// 仅演示用，只关心 indexer，不处理事件，所以传一个空的 HandlerFunc，
	// 实际使用中一般不会这样做
	indexer, informer := cache.NewIndexerInformer(
		lw, &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{}, indexers)

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go informer.Run(stopCh)

	// 在 informer 首次同步完成后再操作
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		panic("timed out waiting for caches to sync")
	}

	// 获取 cache.NamespaceIndex 索引下，索引值为 "tmp" 中的所有键
	keys, err := indexer.IndexKeys(cache.NamespaceIndex, "tmp")
	magicconch.Must(err)
	for _, k := range keys {
		fmt.Println(k)
	}
}
