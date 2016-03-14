package controller

import (
	"github.com/golang/glog"
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
	osclient "github.com/openshift/origin/pkg/client"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
)

// NamespaceController is responsible for participating in Kubernetes Namespace termination
// Use the NamespaceControllerFactory to create this controller.
type BackingServiceController struct {
	// Client is an OpenShift client.
	Client osclient.Interface
	// KubeClient is a Kubernetes client.
	KubeClient kclient.Interface
}

type fatalError string

func (e fatalError) Error() string {
	return "fatal error handling BackingServiceController: " + string(e)
}

// Handle processes a namespace and deletes content in origin if its terminating
func (c *BackingServiceController) Handle(bs *backingserviceapi.BackingService) (err error) {
	glog.Info("bs handle called.")

	if bs.Status.Phase != backingserviceapi.BackingServicePhaseActive {
		bs.Status.Phase = backingserviceapi.BackingServicePhaseActive

		c.Client.BackingServices().Update(bs)
	}

	return nil
}
