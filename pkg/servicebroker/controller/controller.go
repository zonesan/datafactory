package controller

import (
	"fmt"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	osclient "github.com/openshift/origin/pkg/client"
	servicebrokerapi "github.com/openshift/origin/pkg/servicebroker/api"

	backingservice "github.com/openshift/origin/pkg/backingservice/api"
	servicebrokerclient "github.com/openshift/origin/pkg/servicebroker/client"
	"time"
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
func (c *ServiceBrokerController) Handle(sb *servicebrokerapi.ServiceBroker) (err error) {

	if sb.Spec.Url == "" {
		return nil
	}

	services, err := c.ServiceBrokerClient.Catalog(sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password)
	if err != nil {

		fmt.Printf("ServiceBroker %s catalog err %s\n", sb.Name, err.Error())
		sb.Status.Phase = servicebrokerapi.ServiceBrokerFailed
		c.Client.ServiceBrokers().Update(sb)
		time.Sleep(time.Second * 60)

		return nil
	}

	for _, v := range services.Services {
		backingService := &backingservice.BackingService{}
		backingService.Spec = backingservice.BackingServiceSpec(v)
		backingService.Annotations = make(map[string]string)
		backingService.Name = sb.Name
		backingService.GenerateName = v.Name
		backingService.Labels = map[string]string{
			servicebrokerapi.ServiceBrokerLabel: sb.Name,
		}
		if _, err := c.Client.BackingServices().Create(backingService); err == nil {
			sb.Status.Phase = servicebrokerapi.ServiceBrokerActive
			c.Client.ServiceBrokers().Update(sb)
		}
	}

	return nil
}
