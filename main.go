package main

import (
	"fmt"
	"os"

	"github.com/rook/rook-operator/pkg/k8sutil"
	"github.com/rook/rook-operator/pkg/operator"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/runtime/schema"
	"k8s.io/client-go/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	k8s "k8s.io/client-go/kubernetes/typed/core/v1"
)

func main() {
	/*kclient, err := unversioned.NewInCluster()
	if err != nil {
		fmt.Printf("failed to create k8s client. %+v\n", err)
		os.Exit(1)
	}*/

	/*config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("failed to create in cluster config. %+v\n", err)
		os.Exit(1)
	}
	config.GroupVersion = &unversioned.GroupVersion{
		Group:   "rook.io",
		Version: "v1",
	}
	restClient, err := rest.RESTClientFor(config)
	if err != nil {
		fmt.Printf("failed to create rest client. %+v\n", err)
		os.Exit(1)
	}*/

	_, restClient, err := getRESTClients()
	if err != nil {
		fmt.Printf("failed to get clients. %+v\n", err)
	}

	client := k8s.New(restClient)
	op := operator.New(k8sutil.Namespace, client)
	err = op.Run()
	if err != nil {
		fmt.Printf("failed to run operator. %+v\n", err)
		os.Exit(1)
	}
}

func getRESTClients() (*kubernetes.Clientset, *rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	groupversion := schema.GroupVersion{
		Group:   "quantum.com",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				groupversion,
				&api.ListOptions{},
				&api.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(api.Scheme)

	client, err := rest.RESTClientFor(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, client, nil
}
