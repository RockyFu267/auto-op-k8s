package main

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	//读取配置
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		log.Panicln(err)
	}
	config.GroupVersion = &corev1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	config.APIPath = "/api"

	//创建链接
	newClient, err := rest.RESTClientFor(config)
	if err != nil {
		log.Panicln(err)
	}

	//获取数据
	podTest := corev1.Pod{}
	err = newClient.Get().Namespace("default").Resource("pods").Name("redis-cluster-redis-cluster-amd64-5").Do(context.TODO()).Into(&podTest)
	if err != nil {
		log.Panicln(err)
	} else {
		fmt.Println(podTest.Name)
	}
}
