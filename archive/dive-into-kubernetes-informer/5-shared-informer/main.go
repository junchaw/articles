package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func main() {
	fmt.Println("----- 5-shared-informer -----")

	lw := newConfigMapsListerWatcher()
	sharedInformer := cache.NewSharedInformer(lw, &corev1.ConfigMap{}, 0)
	// 添加一个处理程序
	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("created, printing namespace: %s\n", configMap.Namespace)
		},
	})
	// 添加另一个处理程序
	sharedInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("created, printing name: %s\n", configMap.Name)
		},
	})

	stopper := make(chan struct{})
	defer close(stopper)

	fmt.Println("Start syncing....")

	go sharedInformer.Run(stopper)

	<-stopper
}
