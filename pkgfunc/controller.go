package pkgfunc

import (
	"log"
	"reflect"

	AN1 "k8s.io/api/networking/v1"
	APPM1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ICv1 "k8s.io/client-go/informers/core/v1"
	INv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	LCv12 "k8s.io/client-go/listers/core/v1"
	LNv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

// controllerDemo demo
type controllerDemo struct {
	client        kubernetes.Interface
	ingressLister LNv1.IngressLister
	ServiceLister LCv12.ServiceLister
	queue         workqueue.RateLimitingInterface
}

func (c *controllerDemo) Run(stopCH chan struct{}) {
	<-stopCH
}

func NewController(client kubernetes.Interface, svcInformer ICv1.ServiceInformer, ingressInformer INv1.IngressInformer) controllerDemo {
	c := controllerDemo{
		client:        client,
		ingressLister: ingressInformer.Lister(),
		ServiceLister: svcInformer.Lister(),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingressManager"),
	}
	svcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addSvc,
		UpdateFunc: c.updateSvc,
	})
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.deleteIngress,
	})
	return c
}

func (c *controllerDemo) addSvc(obj interface{}) {
	c.enqueue(obj)
}

func (c *controllerDemo) updateSvc(oldObj interface{}, newObj interface{}) {
	//比较annotation
	if reflect.DeepEqual(oldObj, newObj) {
		return
	}
	c.enqueue(oldObj)

}

func (c *controllerDemo) deleteIngress(obj interface{}) {
	ingressTMP := obj.(*AN1.Ingress)
	ownerReference := APPM1.GetControllerOf(ingressTMP)
	if ownerReference == nil {
		return
	}
	if ownerReference.Kind != "Service" {
		return
	}
	c.queue.Add(ingressTMP.Namespace + "/" + ingressTMP.Name)

}

func (c *controllerDemo) enqueue(obj interface{}) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		log.Panicln(err)
		return
	}

	c.queue.Add(key)
}
