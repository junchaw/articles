package main

import (
	"k8s.io/client-go/tools/cache"
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
