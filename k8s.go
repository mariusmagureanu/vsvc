package main

import (
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	v1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	cacheInformerUpdateInterval = 5 * time.Second
	varnishBackendAnnotation    = "varnish.backend"
)

// K8sSession is a type which facilitates the
// interaction with a kubernetes cluster.
type K8sSession struct {
	ConfigFile      string
	client          *kubernetes.Clientset
	serviceInformer v1.ServiceInformer
}

var backendStore *BackendStore

// InitK8s creates a session object and
// initializes its k8s client.
func initK8s(kubeConfigFile string) (*K8sSession, error) {
	var (
		config *rest.Config
		err    error
	)

	if kubeConfigFile != "" {
		Debug("using kubeconfig file:", kubeConfigFile)
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
	} else {
		Debug("using in-cluster config")
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	k8sSession := K8sSession{}

	k8sSession.client, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create kubernetes client. %s", err.Error())
	}

	return &k8sSession, nil
}

// startInformers starts the node informer. One informer per
// application is enough.
func (k *K8sSession) startInformers() {
	informerFactory := informers.NewSharedInformerFactory(k.client, cacheInformerUpdateInterval)

	k.serviceInformer = informerFactory.Core().V1().Services()

	k.serviceInformer.Informer().AddEventHandler(getSvcInformerEventHandler())

	informerFactory.WaitForCacheSync(wait.NeverStop)
	informerFactory.Start(wait.NeverStop)
}

func getSvcInformerEventHandler() cache.ResourceEventHandlerFuncs {
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ssp, ok := obj.(*corev1.Service)
			if ok {
				if backendValue, found := ssp.Annotations[varnishBackendAnnotation]; found {
					Debug("added:", ssp.Name, ssp.Namespace, backendValue)

					b := Backend{Name: ssp.Name, Namespace: ssp.Namespace, Port: strconv.Itoa(int(ssp.Spec.Ports[0].Port))}
					backendStore.add(b)
					err := backendStore.updateVCL()

					if err != nil {
						Error(err)
					}
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldService, okOld := oldObj.(*corev1.Service)
			newService, okNew := newObj.(*corev1.Service)

			if okOld && okNew {
				if oldService.ObjectMeta.ResourceVersion != newService.ObjectMeta.ResourceVersion {
					if backendValue, found := newService.Annotations[varnishBackendAnnotation]; found {
						Debug("updated:", newService.Name, newService.Namespace, backendValue)

						b := Backend{Name: newService.Name, Namespace: newService.Namespace, Port: strconv.Itoa(int(newService.Spec.Ports[0].Port))}
						backendStore.add(b)

						err := backendStore.updateVCL()
						if err != nil {
							Error(err)
						}
					} else if backendValue, found := oldService.Annotations[varnishBackendAnnotation]; found {
						Debug("removed:", newService.Name, newService.Namespace, backendValue)

						b := Backend{Name: newService.Name, Namespace: newService.Namespace, Port: strconv.Itoa(int(oldService.Spec.Ports[0].Port))}
						backendStore.delete(b)

						err := backendStore.updateVCL()
						if err != nil {
							Error(err)
						}
					}
				}
			}
		},
	}

	return handler
}
