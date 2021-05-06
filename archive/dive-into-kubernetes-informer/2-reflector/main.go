package main

import (
	"fmt"
	"github.com/spongeprojects/magicconch"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"time"
)

// newStore 用于创建一个 cache.Store 对象，作为当前资源状态的对象存储
func newStore() cache.Store {
	return cache.NewStore(cache.MetaNamespaceKeyFunc)
}

// newQueue 用于创建一个 cache.Queue 对象，这里实现为 FIFO 先进先出队列，
// 注意在初始化时 store 作为 KnownObjects 参数传入其中，
// 因为在重新同步 (resync) 操作中 Reflector 需要知道当前的资源状态，
// 另外在计算变更 (Delta) 时，也需要对比当前的资源状态。
// 这个 KnownObjects 对队列，以及对 Reflector 都是只读的，用户需要自己维护好 store 的状态。
func newQueue(store cache.Store) cache.Queue {
	return cache.NewDeltaFIFOWithOptions(cache.DeltaFIFOOptions{
		KnownObjects:          store,
		EmitDeltaTypeReplaced: true,
	})
}

// newConfigMapsReflector 用于创建一个 cache.Reflector 对象，
// 当 Reflector 开始运行 (Run) 后，队列中就会推入新收到的事件。
func newConfigMapsReflector(queue cache.Queue) *cache.Reflector {
	lw := newConfigMapsListerWatcher() // 前面有说明
	return cache.NewReflector(lw, &corev1.ConfigMap{}, queue, 0)
}

func main() {
	fmt.Println("----- 2-reflector -----")

	store := newStore()
	queue := newQueue(store)
	reflector := newConfigMapsReflector(queue)

	stopCh := make(chan struct{})
	defer close(stopCh)

	// reflector 开始运行后，队列中就会推入新收到的事件
	go reflector.Run(stopCh)

	// 注意处理事件过程中维护好 store 状态，包括 Add, Update, Delete 操作，
	// 否则会出现不同步问题，在 Informer 当中这些逻辑都已经被封装好了，但目前我们还需要关心一下。
	processObj := func(obj interface{}) error {
		// 最先收到的事件会被最先处理
		for _, d := range obj.(cache.Deltas) {
			switch d.Type {
			case cache.Sync, cache.Replaced, cache.Added, cache.Updated:
				if _, exists, err := store.Get(d.Object); err == nil && exists {
					if err := store.Update(d.Object); err != nil {
						return err
					}
				} else {
					if err := store.Add(d.Object); err != nil {
						return err
					}
				}
			case cache.Deleted:
				if err := store.Delete(d.Object); err != nil {
					return err
				}
			}
			configMap, ok := d.Object.(*corev1.ConfigMap)
			if !ok {
				return fmt.Errorf("not config: %T", d.Object)
			}
			fmt.Printf("%s: %s\n", d.Type, configMap.Name)
		}
		return nil
	}

	fmt.Println("Start syncing...")

	// 持续运行直到 stopCh 关闭
	wait.Until(func() {
		for {
			_, err := queue.Pop(processObj)
			magicconch.Must(err)
		}
	}, time.Second, stopCh)
}
