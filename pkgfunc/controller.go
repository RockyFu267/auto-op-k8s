package pkgfunc

import (
	"context"
	"reflect"
	"time"

	ACV1 "k8s.io/api/core/v1"
	AN1 "k8s.io/api/networking/v1"
	KAPAE "k8s.io/apimachinery/pkg/api/errors"
	APAM1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	KAPPUR "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	CIC1 "k8s.io/client-go/informers/core/v1"
	CIN1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	CLC1 "k8s.io/client-go/listers/core/v1"
	CLN1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	workNum  = 5
	maxRetry = 10
)

// controllerDemo demo
type controllerDemo struct {
	client        kubernetes.Interface
	ingressLister CLN1.IngressLister
	serviceLister CLC1.ServiceLister
	queue         workqueue.RateLimitingInterface
}

func (c *controllerDemo) Run(stopCH chan struct{}) {
	for i := 0; i < workNum; i++ {
		go wait.Until(c.worker, time.Minute, stopCH)
	}
	<-stopCH
}

func NewController(client kubernetes.Interface, svcInformer CIC1.ServiceInformer, ingressInformer CIN1.IngressInformer) controllerDemo {
	c := controllerDemo{
		client:        client,
		ingressLister: ingressInformer.Lister(),
		serviceLister: svcInformer.Lister(),
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
	ownerReference := APAM1.GetControllerOf(ingressTMP)
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
		KAPPUR.HandleError(err)
		return
	}

	c.queue.Add(key)
}

func (c *controllerDemo) worker() {
	for c.processNextItem() {

	}
}

func (c *controllerDemo) processNextItem() bool {
	item, shutdown := c.queue.Get()
	if shutdown {
		return false
	}
	defer c.queue.Done(item)

	key := item.(string)

	err := c.syncService(key)
	if err != nil {
		c.handlerError(key, err)
	}
	return true
}
func (c *controllerDemo) syncService(key string) error {
	namespaceKey, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	//删除
	service, err := c.serviceLister.Services(namespaceKey).Get(name)
	if KAPAE.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	//新增和删除
	_, ok := service.GetAnnotations()["ingress/http"]
	ingress, err := c.ingressLister.Ingresses(namespaceKey).Get(name)
	if err != nil && !KAPAE.IsNotFound(err) {
		return err
	}

	if ok && KAPAE.IsNotFound(err) {
		//create ingress
		ig := c.constructIngress(service)
		_, err := c.client.NetworkingV1().Ingresses(namespaceKey).Create(context.TODO(), ig, APAM1.CreateOptions{})
		if err != nil {
			return err
		}
	} else if !ok && ingress != nil {
		//delete ingress
		err := c.client.NetworkingV1().Ingresses(namespaceKey).Delete(context.TODO(), name, APAM1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *controllerDemo) handlerError(key string, err error) {
	if c.queue.NumRequeues(key) <= maxRetry {
		c.queue.AddRateLimited(key)
		return
	}

	KAPPUR.HandleError(err)
	c.queue.Forget(key)
}

func (c *controllerDemo) constructIngress(service *ACV1.Service) *AN1.Ingress {
	ingress := AN1.Ingress{}

	ingress.ObjectMeta.OwnerReferences = []APAM1.OwnerReference{
		*APAM1.NewControllerRef(service, ACV1.SchemeGroupVersion.WithKind("Service")),
	}

	ingress.Name = service.Name
	ingress.Namespace = service.Namespace
	pathType := AN1.PathTypePrefix
	icn := "nginx"
	ingress.Spec = AN1.IngressSpec{
		IngressClassName: &icn,
		Rules: []AN1.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: AN1.IngressRuleValue{
					HTTP: &AN1.HTTPIngressRuleValue{
						Paths: []AN1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: AN1.IngressBackend{
									Service: &AN1.IngressServiceBackend{
										Name: service.Name,
										Port: AN1.ServiceBackendPort{
											Number: 80,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return &ingress
}
