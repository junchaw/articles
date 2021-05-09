package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	v1 "github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/api/stable.wbsnail.com/v1"
	"github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client/clientset/versioned"
	"github.com/wbsnail/articles/archive/dive-into-kubernetes-informer/11-crd-informer/client/informers/externalversions"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func main() {
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	magicconch.Must(err)

	clientset, err := versioned.NewForConfig(config)
	magicconch.Must(err)

	informerFactory := externalversions.NewSharedInformerFactory(clientset, 0)
	rabbitInformer := informerFactory.Stable().V1().Rabbits().Informer()
	rabbitInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			rabbit, ok := obj.(*v1.Rabbit)
			if !ok {
				return
			}
			fmt.Printf("A rabbit is created: %s\n", rabbit.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newRabbit, ok := oldObj.(*v1.Rabbit)
			if !ok {
				return
			}
			fmt.Printf("A rabbit is updated: %s\n", newRabbit.Name)
		},
		DeleteFunc: func(obj interface{}) {
			rabbit, ok := obj.(*v1.Rabbit)
			if !ok {
				return
			}
			fmt.Printf("A rabbit is deleted: %s\n", rabbit.Name)
		},
	})

	stopCh := make(chan struct{})
	defer close(stopCh)

	fmt.Println("Start syncing....")

	go informerFactory.Start(stopCh)

	<-stopCh
}
