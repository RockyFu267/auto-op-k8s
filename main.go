package main

import (
	"context"
	"fmt"
	"log"

	MetaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// //读取配置
	// config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	// if err != nil {
	// 	log.Panicln(err)
	// }
	// config.GroupVersion = &corev1.SchemeGroupVersion
	// config.NegotiatedSerializer = scheme.Codecs
	// config.APIPath = "/api"

	// //创建链接
	// newClient, err := rest.RESTClientFor(config)
	// if err != nil {
	// 	log.Panicln(err)
	// }

	// //获取数据
	// podTest := corev1.Pod{}
	// err = newClient.Get().Namespace("default").Resource("pods").Name("redis-cluster-redis-cluster-amd64-5").Do(context.TODO()).Into(&podTest)
	// if err != nil {
	// 	log.Panicln(err)
	// } else {
	// 	fmt.Println(podTest.Name)
	// }

	//读取配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Println(err)
	}
	newClent, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Println(err)
	}
	coreV1 := newClent.CoreV1()
	podTest, err := coreV1.Pods("default").Get(context.TODO(), "redis-cluster-redis-cluster-amd64-7", MetaV1.GetOptions{})
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(podTest.Name)
	}

}
