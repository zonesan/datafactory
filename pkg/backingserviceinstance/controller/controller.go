package controller

import (
	backingserviceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	osclient "github.com/openshift/origin/pkg/client"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type BackingServiceInstanceController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling BackingServiceInstanceController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *BackingServiceInstanceController) Handle(bs *backingserviceapi.BackingServiceInstance) (err error) {

	if bs.Status != backingserviceapi.BackingServiceInstanceStatusActive {
		bs.Status = backingserviceapi.BackingServiceInstanceStatusActive

		c.Client.BackingServiceInstances().Update(bs)
	}

	return nil
}
