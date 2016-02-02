package controller

import (
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"
	osclient "github.com/openshift/origin/pkg/client"

	servicebrokerclient "github.com/openshift/origin/pkg/servicebroker/client"
	"fmt"
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

	b, err := c.ServiceBrokerClient.Catalog(bs.Spec.Url);
	if err != nil {
		fmt.Printf("servicebroker controller catalog err %s", err.Error())
		return err
	}
	fmt.Printf("[Debug] -------------------> %#v \n", bs)
	fmt.Printf("[Debug] -------------------> %s \n", string(b))

	//if bs.Status.Phase != "test" {
	//	bs.Status.Phase = "test"
	//
	//	c.Client.ServiceBrokers().Update(bs)
	//}

	return nil
}
