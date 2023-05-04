package pkgfunc

import (
	ICv1 "k8s.io/client-go/informers/core/v1"
	INv1 "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	LCv12 "k8s.io/client-go/listers/core/v1"
	LNv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

// controllerDemo demo
type controllerDemo struct {
	client        kubernetes.Interface
	ingressLister LNv1.IngressLister
	ServiceLister LCv12.ServiceLister
}

func (c *controllerDemo) Run(stopCH chan struct{}) {
	<-stopCH
}

func NewController(client kubernetes.Interface, svcInformer ICv1.ServiceInformer, ingressInformer INv1.IngressInformer) controllerDemo {
	c := controllerDemo{
		client:        client,
		ingressLister: ingressInformer.Lister(),
		ServiceLister: svcInformer.Lister(),
	}
	svcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		AddFunc:    c.addSvc,
		UpdateFunc: c.updateSvc,
	})
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerDetailedFuncs{
		DeleteFunc: c.deleteIngress,
	})
	return c
}

func (c *controllerDemo) addSvc(obj interface{}, objBool bool) {

}

func (c *controllerDemo) updateSvc(oldObj interface{}, newObj interface{}) {

}

func (c *controllerDemo) deleteIngress(obj interface{}) {

}
