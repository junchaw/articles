package main

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

func newController() cache.Controller {
	lw := newConfigMapsListerWatcher()
	store := newStore()
	queue := newQueue(store)
	cfg := &cache.Config{
		Queue:            queue,
		ListerWatcher:    lw,
		ObjectType:       &corev1.ConfigMap{},
		FullResyncPeriod: 0,
		RetryOnError:     false,
		Process: func(obj interface{}) error {
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
		},
	}
	return cache.New(cfg)
}

func main() {
	controller := newController()

	stopper := make(chan struct{})
	defer close(stopper)

	fmt.Println("Start syncing....")

	go controller.Run(stopper)

	<-stopper
}
