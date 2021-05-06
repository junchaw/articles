package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func main() {
	fmt.Println("----- 9-shared-informer-factory -----")

	// mustClientset 用于创建 kubernetes.Interface 实例；
	// 第 2 个参数是 defaultResync，是构建新 Informer 时默认的 resyncPeriod，
	// resyncPeriod 在前一部分中介绍过了；
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		mustClientset(), 0, informers.WithNamespace("tmp"))
	configMapsInformer := informerFactory.Core().V1().ConfigMaps().Informer()
	configMapsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			configMap, ok := obj.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("created: %s\n", configMap.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			configMap, ok := newObj.(*corev1.ConfigMap)
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

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go configMapsInformer.Run(stopCh)

	<-stopCh
}
