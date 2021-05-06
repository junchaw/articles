package main

import (
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// newConfigMapsListerWatcher 用于创建 tmp namespace 下 configmaps 资源的 ListerWatcher 实例
func newConfigMapsListerWatcher() cache.ListerWatcher {
	clientset := mustClientset()              // 前面有说明
	client := clientset.CoreV1().RESTClient() // 客户端，请求器
	resource := "configmaps"                  // GET 请求参数之一
	namespace := "tmp"                        // GET 请求参数之一
	selector := fields.Everything()           // GET 请求参数之一
	lw := cache.NewListWatchFromClient(client, resource, namespace, selector)
	return lw
}
