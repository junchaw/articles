package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("resource missing")
		return
	}
	// 资源，比如 "configmaps.v1.", "deployments.v1.apps", "rabbits.v1.stable.wbsnail.com"
	resource := os.Args[1]

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	magicconch.Must(err)

	// 注意创建了 dynamicClient, 而不是 clientset
	dynamicClient, err := dynamic.NewForConfig(config)
	magicconch.Must(err)

	// 同样这里也是 DynamicSharedInformerFactory, 而不是 SharedInformerFactory
	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		dynamicClient, 0, "tmp", nil)

	// 通过 schema 包提供的 ParseResourceArg 由资源描述字符串解析出 GroupVersionResource
	gvr, _ := schema.ParseResourceArg(resource)
	if gvr == nil {
		fmt.Println("cannot parse gvr")
		return
	}
	// 使用 gvr 动态生成 Informer
	informer := informerFactory.ForResource(*gvr).Informer()
	// 熟悉的代码，熟悉的味道，只是收到的 obj 类型好像不太一样
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			// *unstructured.Unstructured 类是所有 Kubernetes 资源类型公共方法的抽象，
			// 提供所有对公共属性的访问方法，像 GetName, GetNamespace, GetLabels 等等，
			s, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return
			}
			fmt.Printf("created: %s\n", s.GetName())
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldS, ok1 := oldObj.(*unstructured.Unstructured)
			newS, ok2 := newObj.(*unstructured.Unstructured)
			if !ok1 || !ok2 {
				return
			}
			// 要访问公共属性外的字段，可以借助 unstructured 包提供的一些助手方法：
			oldColor, ok1, err1 := unstructured.NestedString(oldS.Object, "spec", "color")
			newColor, ok2, err2 := unstructured.NestedString(newS.Object, "spec", "color")
			if !ok1 || !ok2 || err1 != nil || err2 != nil {
				fmt.Printf("updated: %s\n", newS.GetName())
			}
			fmt.Printf("updated: %s, old color: %s, new color: %s\n", newS.GetName(), oldColor, newColor)
		},
		DeleteFunc: func(obj interface{}) {
			s, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return
			}
			fmt.Printf("deleted: %s\n", s.GetName())
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go informerFactory.Start(stopCh)

	<-stopCh
}
