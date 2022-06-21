package main

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			AddFunc:    c.handleAdd,
			DeleteFunc: c.handleDelete,
		},
	)

	return c
}

func (c *controller) run(ch <-chan struct{}) {
	fmt.Println("Initializing controller...")
	if !cache.WaitForCacheSync(ch, c.depCacheSynced) {
		fmt.Println("Waiting cache syncronization...")
	}

	//make the call to worker every 1 second
	go wait.Until(c.worker, 1*time.Second, ch)

	<-ch

}

//function that continuosly fetch the values in the queue and then run the logic
func (c *controller) worker() {
	for c.processItem() {

	}
}

func (c *controller) processItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false //no object in queue
	}

	key, err := cache.MetaNamespaceKeyFunc(item)
	if err != nil {
		fmt.Printf("Getting keu from cache %s\n", err.Error())
	}

	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		fmt.Printf("Splitting key into ns and name %s\n", err.Error())
		return false
	}

	err = c.syncDeployment(ns, name)
	if err != nil {
		fmt.Printf("Syncing deployment %s\n", err.Error())
		return false
	}

	return true

}

//Create a service and ingress
func (c *controller) syncDeployment(ns, name string) error {

	ctx := context.Background()
	svc := corev1.Service{}
	_, err := c.clientset.CoreV1().Services(ns).Create(ctx, &svc, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("SCreating service %s\n", err.Error())
	}

	return nil
}

func (c *controller) handleAdd(obj interface{}) {
	println("Testing handleAdd func")
	c.queue.Add(obj) //When a new deployment happens add an object in queue

}

func (c *controller) handleDelete(obj interface{}) {
	println("Testing handleDelete func")
	c.queue.Add(obj)
}
