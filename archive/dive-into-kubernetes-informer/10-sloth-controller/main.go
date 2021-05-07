package main

import (
	"flag"
	"fmt"
	"github.com/spongeprojects/magicconch"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"os"
	"os/signal"
	"time"
)

// SlothController 树懒控制器!
type SlothController struct {
	factory           informers.SharedInformerFactory
	configMapLister   listerscorev1.ConfigMapLister
	configMapInformer cache.SharedIndexInformer
	queue             workqueue.RateLimitingInterface

	// maxRetries 树懒君需要重试多少次才会放弃
	maxRetries int
	// chanceOfFailure 树懒君处理任务有多少概率失败（百分比）
	chanceOfFailure int
	// nap 树懒君睡一觉要多久
	nap time.Duration
}

func NewController(
	factory informers.SharedInformerFactory,
	queue workqueue.RateLimitingInterface,
	chanceOfFailure int,
	nap time.Duration,
) *SlothController {
	configMapLister := factory.Core().V1().ConfigMaps().Lister()
	configMapInformer := factory.Core().V1().ConfigMaps().Informer()
	configMapInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old interface{}, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				// 只简单地把 key 放到队列中
				klog.Infof("[%s] received", key)
				queue.Add(key)
			}
		},
	})

	return &SlothController{
		factory:           factory,
		configMapLister:   configMapLister,
		configMapInformer: configMapInformer,
		queue:             queue,
		maxRetries:        3,
		chanceOfFailure:   chanceOfFailure,
		nap:               nap,
	}
}

// Run 开始运行控制器直到出错或 stopCh 关闭
func (c *SlothController) Run(sloths int, stopCh chan struct{}) error {
	// runtime.HandleCrash 是 Kubernetes 官方提供的 panic recover 方法，
	// 提供一个 panic recover 的统一入口，
	// 默认只是记录日志，该 panic 还是 panic。
	defer runtime.HandleCrash()
	// 关闭队列让 sloths 不要再处理任务
	defer c.queue.ShutDown()

	klog.Info("SlothController starting...")

	go c.factory.Start(stopCh)

	// 等待首次同步完成
	for _, ok := range c.factory.WaitForCacheSync(stopCh) {
		if !ok {
			return fmt.Errorf("timed out waiting for caches to sync")
		}
	}

	for i := 0; i < sloths; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	// 等待 stopCh 关闭
	<-stopCh

	return nil
}

func (c *SlothController) runWorker() {
	for c.processNextItem() {
	}
}

// processNextItem 用于等待和处理队列中的新任务
func (c *SlothController) processNextItem() bool {
	// 阻塞住，等待新任务中...
	key, shutdown := c.queue.Get()
	if shutdown {
		return false // 队列已进入退出状态，不要继续处理
	}

	// 任务完成后记得标记完成，因为尽管有多个 sloths，
	// 但对相同 key 的多个任务是不会并行处理的。
	defer c.queue.Done(key)

	result := c.processItem(key)
	c.handleErr(key, result)

	return true
}

// processItem 用于同步处理一个任务
func (c *SlothController) processItem(key interface{}) error {
	// 处理任务很慢，因为树懒很懒
	time.Sleep(c.nap)

	if rand.Intn(100) < c.chanceOfFailure {
		// 睡过头啦！
		return fmt.Errorf("zzz... ")
	}

	klog.Infof("[%s] processed.", key)
	return nil
}

// handleErr 用于检查任务处理结果，在必要的时候重试
func (c *SlothController) handleErr(key interface{}, result error) {
	if result == nil {
		// 每次执行成功后清空重试记录。
		c.queue.Forget(key)
		return
	}

	if c.queue.NumRequeues(key) < c.maxRetries {
		klog.Warningf("error processing %s: %v", key, result)
		// 重试
		c.queue.AddRateLimited(key)
		return
	}

	// 重试次数过多，日志记录错误，同时也别忘了清空重试记录
	c.queue.Forget(key)
	// runtime.HandleError 是 Kubernetes 官方提供的错误处理错误响应方法，
	// 提供一个错误响应的统一入口。
	runtime.HandleError(fmt.Errorf("max retries exceeded, "+
		"dropping item %s out of the queue: %v", key, result))
}

func main() {
	fmt.Println("----- 10-sloth-controller -----")

	var sloths int
	var chanceOfFailure int
	var napInSecond int
	flag.IntVar(&sloths, "sloths", 1,
		"number of sloths")
	flag.IntVar(&chanceOfFailure, "chance-of-failure", 0,
		"chance of failure in percentage")
	flag.IntVar(&napInSecond, "nap", 0,
		"how long should the sloth nap (in seconds)")
	flag.Parse()

	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	magicconch.Must(err)

	clientset, err := kubernetes.NewForConfig(config)
	magicconch.Must(err)

	// 创建 SharedInformerFactory
	defaultResyncPeriod := time.Hour * 12
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset, defaultResyncPeriod, informers.WithNamespace("tmp"))

	// 使用默认配置创建 RateLimitingQueue，这种队列支持重试，同时会记录重试次数
	rateLimiter := workqueue.DefaultControllerRateLimiter()
	queue := workqueue.NewRateLimitingQueue(rateLimiter)

	controller := NewController(informerFactory, queue, chanceOfFailure,
		time.Duration(napInSecond)*time.Second)

	stopCh := make(chan struct{})

	// 响应中断信号
	interrupted := make(chan os.Signal)
	signal.Notify(interrupted, os.Interrupt)

	// 当 interrupted 关闭时，关闭 stopCh
	go func() {
		<-interrupted
		close(stopCh)
	}()

	if err := controller.Run(sloths, stopCh); err != nil {
		klog.Errorf("SlothController exit with error: %s", err)
	} else {
		klog.Info("SlothController exit")
	}
}
