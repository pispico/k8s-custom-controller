package main

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	applisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type controller struct {
	clientset      kubernetes.Interface
	depLister      applisters.DeploymentLister //list Deployments
	depCacheSynced cache.InformerSynced        //function that can be used to determine if an informer has synced
	queue          workqueue.RateLimitingInterface
}

const (
	queuename = "cc-deploy-queue"
)

func NewController(clientset kubernetes.Interface, depInformer appsinformers.DeploymentInformer) *controller {
	c := &controller{
		clientset:      clientset,
		depLister:      depInformer.Lister(),
		depCacheSynced: depInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), queuename),
	}

	depInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    handleAdd,
			DeleteFunc: handleDelete,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Initializing controller...")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
		fmt.Println("Waiting cache syncronization...")
	}

	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch

}

func (c *controller) worker() {

}

func handleAdd(obj interface{}) {
	println("Testing handleAdd func")
}

func handleDelete(obj interface{}) {
	println("Testing handleDelete func")
}
