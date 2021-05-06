package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func main() {
	fmt.Println("----- 4-informer -----")

	lw := newConfigMapsListerWatcher()
	// 第一个返回的参数是 cache.Store，这里暂时用不到所以直接丢弃
	_, controller := cache.NewInformer(lw, &corev1.ConfigMap{}, 0, cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("created: %s\n", configMap.Name)
		},
		UpdateFunc: func(old, new interface{}) {
			configMap, ok := old.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("updated: %s\n", configMap.Name)
		},
		DeleteFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("deleted: %s\n", configMap.Name)
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	fmt.Println("Start syncing....")

	go controller.Run(stopper)

	<-stopper
}
