/*
Copyright 2016 Intel Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"time"

	//"kube-crd/client"
	//"kube-crd/crd"
	"citcrd/cit_tl"
	"citcrd/tlclient"
	//"k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/fields"

	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	//apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	//apierrors "k8s.io/apimachinery/pkg/api/errors"
	//meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	//"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	//informers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	//"k8s.io/apiextensions-apiserver/pkg/client/clientset/internalclientset"
	"flag"
)

// return rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {

	kubeconf := flag.String("kubeconf", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := GetClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// Create a new clientset which include our CRD schema
        crdcs, scheme, err := tlclient.NewClient(config)
        if err != nil {
                panic(err)
        }

	// Create a CRD client interface
        tlcrdclient := tlclient.CitTLClient(crdcs, scheme, "default")

	cs, err := apiextcs.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// note: if the CRD exist our CreateCRD function is set to exit without an error
	err = cit_tl.NewcitTLCustomResourceDefinition(cs)
	if err != nil {
		panic(err)
	}

	// Wait for the CRD to be created before we use it (only needed if its a new one)
	time.Sleep(3 * time.Second)

	// Create a queue 
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "citTLcontroller")

	
	indexer, informer := cache.NewIndexerInformer(tlcrdclient.NewListWatch(), &tlclient.Trusttab{}, 0, cache.ResourceEventHandlerFuncs{
                AddFunc: func(obj interface{}) {
                        key, err := cache.MetaNamespaceKeyFunc(obj)
			fmt.Println("Received Add event f0r %#v", key)
                        if err == nil {
                                queue.Add(key)
                        }
                },
                UpdateFunc: func(old interface{}, new interface{}) {
                        key, err := cache.MetaNamespaceKeyFunc(new)
			fmt.Println("Received Update event f0r %#v", key)
                        if err == nil {
                                queue.Add(key)
                        }
                },
                DeleteFunc: func(obj interface{}) {
                        // IndexerInformer uses a delta queue, therefore for deletes we have to use this
                        // key function.
                        key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			fmt.Println("Received delete event f0r %#v", key)
                        if err == nil {
                                queue.Add(key)
                        }
                },
        }, cache.Indexers{})
	
	controller := cit_tl.NewCitTLController(queue, indexer, informer)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(1, stop)

	// Wait forever
	select {}
}
