package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
	"os"
	"os/signal"
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

func main() {
	fmt.Println("----- 1-list-watcher -----")

	lw := newConfigMapsListerWatcher()

	// list 的类型为 runtime.Object, 需要经过反射或类型转换才能使用，
	// 传入的 ListOptions 中的 FieldSelector 始终会被替换为前面的 selector
	list, err := lw.List(metav1.ListOptions{})
	magicconch.Must(err)

	// meta 包封装了一些处理 runtime.Object 对象的方法，屏蔽了反射和类型转换的过程，
	// 提取出的 items 类型为 []runtime.Object
	items, err := meta.ExtractList(list)
	magicconch.Must(err)

	fmt.Println("Initial list:")

	for _, item := range items {
		configMap, ok := item.(*corev1.ConfigMap)
		if !ok {
			return
		}
		fmt.Println(configMap.Name)

		// 如果只关注 meta 信息，也可以使用 meta.Accessor 方法
		// accessor, err := meta.Accessor(item)
		// magicconch.Must(err)
		// fmt.Println(accessor.GetName())
	}

	listMetaInterface, err := meta.ListAccessor(list)
	magicconch.Must(err)

	// resourceVersion 在同步过程中非常重要，看下面它在 Watch 接口中的使用
	resourceVersion := listMetaInterface.GetResourceVersion()

	// w 的类型为 watch.Interface，提供 ResultChan 方法读取事件，
	// 和 List 一样，传入的 ListOptions 中的 FieldSelector 始终会被替换为前面的 selector，
	// ResourceVersion 是 Watch 时非常重要的参数，
	// 它代表一次客户端与服务器进行交互时对应的资源版本，
	// 结合另一个参数 ResourceVersionMatch，表示本次请求对 ResourceVersion 的筛选，
	// 比如以下请求表示：获取版本新于 resourceVersion 的事件。
	// 在考虑连接中断和定期重新同步 (resync) 的情况下，
	// 对 ResourceVersion 的管理就变得更为复杂，我们先不考虑这些情况。
	w, err := lw.Watch(metav1.ListOptions{
		ResourceVersion: resourceVersion,
	})
	magicconch.Must(err)

	stopCh := make(chan os.Signal)
	signal.Notify(stopCh, os.Interrupt)

	fmt.Println("Start watching...")

loop:
	for {
		select {
		case <-stopCh:
			fmt.Println("Interrupted")
			break loop
		case event, ok := <-w.ResultChan():
			if !ok {
				fmt.Println("Broken channel")
				break loop
			}
			configMap, ok := event.Object.(*corev1.ConfigMap)
			if !ok {
				return
			}
			fmt.Printf("%s: %s\n", event.Type, configMap.Name)
		}
	}
}
