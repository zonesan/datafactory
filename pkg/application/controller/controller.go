package controller

import (
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	applicationapi "github.com/openshift/origin/pkg/application/api"
	osclient "github.com/openshift/origin/pkg/client"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type ApplicationController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling ApplicationController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *ApplicationController) Handle(sb *applicationapi.Application) (err error) {

	//if sb.Spec.Url == "" {
	//	return nil
	//}
	//
	//services, err := c.ApplicationClient.Catalog(sb.Spec.Url)
	//if err != nil {
	//
	//	fmt.Printf("Application %s catalog err %s\n", sb.Name, err.Error())
	//	sb.Status.Phase = applicationapi.ApplicationFailed
	//	c.Client.Applications().Update(sb)
	//	time.Sleep(time.Second * 60)
	//
	//	return nil
	//}
	//
	//for _, v := range services.Services {
	//	backingService := &backingservice.BackingService{}
	//	backingService.Spec = backingservice.BackingServiceSpec(v)
	//	backingService.Annotations = make(map[string]string)
	//	backingService.Name = sb.Name
	//	backingService.GenerateName = v.Name
	//	backingService.Labels = map[string]string{
	//		applicationapi.ApplicationLabel: sb.Name,
	//	}
	//	if _, err := c.Client.BackingServices().Create(backingService); err == nil {
	//		sb.Status.Phase = applicationapi.ApplicationActive
	//		c.Client.Applications().Update(sb)
	//	}
	//}

	return nil
}
