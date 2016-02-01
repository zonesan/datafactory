package controller

import (
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"
	osclient "github.com/openshift/origin/pkg/client"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type ServiceBrokerController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling ServiceBrokerController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *ServiceBrokerController) Handle(bs *servicebrokerapi.ServiceBroker) (err error) {


	if bs.Status.Phase != "test" {
		bs.Status.Phase = "test"

		c.Client.ServiceBrokers().Update(bs)
	}

	return nil
}
