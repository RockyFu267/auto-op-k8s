package main

import (
	//1. config
	//2. client
	//3. informer
	//4. add envent handler
	//5. informer.start

	"autp-op-k8s/pkgfunc"
	"log"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	CGrest "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	configNew, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		inClusterConfig, err := CGrest.InClusterConfig()
		if err != nil {
			log.Println("get config error:", err)
		}
		configNew = inClusterConfig
	}

	clinetNew, err := kubernetes.NewForConfig(configNew)
	if err != nil {
		log.Println("create client error:", err)
	}

	factoryNew := informers.NewSharedInformerFactory(clinetNew, 0)
	svcInformer := factoryNew.Core().V1().Services()
	ingressInformer := factoryNew.Networking().V1().Ingresses()

	controllerNew := pkgfunc.NewController(clinetNew, svcInformer, ingressInformer)
	stopCH := make(chan struct{})
	factoryNew.Start(stopCH)
	factoryNew.WaitForCacheSync(stopCH)

	controllerNew.Run(stopCH)

}
