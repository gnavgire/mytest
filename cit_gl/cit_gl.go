package cit_gl

import (
         "fmt"
         "time"

        "github.com/golang/glog"

        //"k8s.io/apimachinery/pkg/runtime/schema"
        // utilruntime "k8s.io/apimachinery/pkg/util/runtime"
        "k8s.io/apimachinery/pkg/util/wait"
        "k8s.io/apimachinery/pkg/util/runtime"
        "k8s.io/client-go/util/workqueue"
        clientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
        "k8s.io/client-go/tools/cache"
        // "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
        // "k8s.io/apiextensions-apiserver/pkg/client/informers/internalversion/apiextensions/internalversion"
        //listers "k8s.io/apiextensions-apiserver/pkg/client/listers/apiextensions/internalversion"
        apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        // internalinterfaces "k8s.io/apiextensions-apiserver/pkg/client/informers/internalversion/internalinterfaces"
	)


type citGLController struct {
        indexer  cache.Indexer
        informer cache.Controller
        queue workqueue.RateLimitingInterface
}

//func NewcitGLCustomResourceDefinition(cs clientset.Interface) *apiextensionsv1beta1.CustomResourceDefinition {
func NewcitGLCustomResourceDefinition(cs clientset.Interface) error {
        crd := &apiextensionsv1beta1.CustomResourceDefinition{
                ObjectMeta: metav1.ObjectMeta{Name: "geotabs.cit.intel.com"},
                Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
                        Group:   "cit.intel.com",
                        Version: "v1beta1",
                        Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
                                Plural:     "geotabs",
                                Singular:   "geotab",
                                Kind:       "GeoTab",
                                ShortNames:  []string{"gt" },
                                ListKind:   "geotabList",
                        },
                        Scope: apiextensionsv1beta1.NamespaceScoped,
                },
        }
        _, err := cs.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
        if err != nil && apierrors.IsAlreadyExists(err) {
                fmt.Println("Cit GL CRD allready exsists")
                return nil
        }
        fmt.Println("sucessfully created CITGL CRD")
        return err
}

func NewCitGLController(queue workqueue.RateLimitingInterface, indexer cache.Indexer, informer cache.Controller) *citGLController {
        return &citGLController{
                informer: informer,
                indexer:  indexer,
                queue:    queue,
        }
}

func (c *citGLController) processNextItem() bool {
        // Wait until there is a new item in the working queue
        key, quit := c.queue.Get()
        if quit {
                return false
        }
        // Tell the queue that we are done with processing this key. This unblocks the key for other workers
        // This allows safe parallel processing because two pods with the same key are never processed in
        // parallel.
        defer c.queue.Done(key)

        // Invoke the method containing the business logic
        err := c.syncToStdout(key.(string))
        // Handle the error if something went wrong during the execution of the business logic
        c.handleErr(err, key)
        return true
}

// syncToStdout is the business logic of the controller. In this controller it simply prints
// information about the pod to stdout. In case an error happened, it has to simply return the error.
// The retry logic should not be part of the business logic.
func (c *citTLController) syncToStdout(key string) error {
        obj, exists, err := c.indexer.GetByKey(key)
        if err != nil {
                glog.Errorf("Fetching object with key %s from store failed with %v", key, err)
                return err
        }

        if !exists {
                // Below we will warm up our cache with a Pod, so that we will see a delete for one pod
                fmt.Printf("Pod %s does not exist anymore\n", key)
        } else {
                // Note that you also have to check the uid if you have a local controlled resource, which
                // is dependent on the actual instance, to detect that a Pod was recreated with the same name
                //fmt.Printf("Sync/Add/Update for Pod %s\n", obj.(*v1.Pod).GetName())
                fmt.Printf("Sync/Add/Update for Pod %#v \n", obj)
        }
        return nil
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *citGLController) handleErr(err error, key interface{}) {
        if err == nil {
                // Forget about the #AddRateLimited history of the key on every successful synchronization.
                // This ensures that future processing of updates for this key is not delayed because of
                // an outdated error history.
                c.queue.Forget(key)
                return
        }

        // This controller retries 5 times if something goes wrong. After that, it stops trying.
        if c.queue.NumRequeues(key) < 5 {
                glog.Infof("Error syncing pod %v: %v", key, err)

                // Re-enqueue the key rate limited. Based on the rate limiter on the
                // queue and the re-enqueue history, the key will be processed later again.
                c.queue.AddRateLimited(key)
                return
        }

        c.queue.Forget(key)
        // Report to an external entity that, even after several retries, we could not successfully process this key
        runtime.HandleError(err)
        glog.Infof("Dropping pod %q out of the queue: %v", key, err)
}

func (c *citGLController) Run(threadiness int, stopCh chan struct{}) {
        defer runtime.HandleCrash()

        // Let the workers stop when we are done
        defer c.queue.ShutDown()
        glog.Info("Starting Pod controller")

        go c.informer.Run(stopCh)

        // Wait for all involved caches to be synced, before processing items from the queue is started
        if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
                runtime.HandleError(fmt.Errorf("Timed out waiting for caches to sync"))
                return
        }

        for i := 0; i < threadiness; i++ {
                go wait.Until(c.runWorker, time.Second, stopCh)
        }

        <-stopCh
        glog.Info("Stopping Pod controller")
}

func (c *citGLController) runWorker() {
        for c.processNextItem() {
        }
}

