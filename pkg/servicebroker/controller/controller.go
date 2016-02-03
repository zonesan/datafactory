package controller

import (
	"fmt"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	osclient "github.com/openshift/origin/pkg/client"
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"

	backingservice "github.com/openshift/origin/pkg/backingservice/api"
	servicebrokerclient "github.com/openshift/origin/pkg/servicebroker/client"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type ServiceBrokerController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
	//ServiceBrokerClient is a ServiceBroker client
	ServiceBrokerClient servicebrokerclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling ServiceBrokerController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *ServiceBrokerController) Handle(bs *servicebrokerapi.ServiceBroker) (err error) {

	if bs.Spec.Url == "" {
		return nil
	}

	services, err := c.ServiceBrokerClient.Catalog(bs.Spec.Url)
	if err != nil {
		fmt.Printf("servicebroker controller catalog err %s", err.Error())
	}

	for _, v := range services.Services {
		fmt.Printf("---------------------->[Debug] %#v", v)
		backingService := &backingservice.BackingService{}
		backingService.Spec = backingservice.BackingServiceSpec(v)
		backingService.Annotations = make(map[string]string)
		backingService.Name = v.Name
		backingService.GenerateName = v.Name

		c.Client.BackingServices().Create(backingService)
	}

	return nil
}
