package controller

import (
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	osclient "github.com/openshift/origin/pkg/client"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	"fmt"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
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
func (c *BackingServiceInstanceController) Handle(bsi *backingserviceinstanceapi.BackingServiceInstance) (err error) {
	fmt.Println("bsi handler called.")
	if bsi.Status.Phase == backingserviceinstanceapi.BackingServiceInstancePhaseReady {
		return nil
	}

	if ok, bs, err := checkIfPlanidExist(c.Client, bsi.Spec.BackingServicePlanGuid); !ok {
		if bsi.Status.Phase != backingserviceinstanceapi.BackingServiceInstancePhaseError {
			bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
			c.Client.BackingServiceInstances().Update(bsi)
		}

		return err
	} else {
		//bsi.Spec.BackingServiceGuid = string(bs.ObjectMeta.UID)
		bsi.Spec.BackingServiceName = bs.ObjectMeta.Name
	}

	bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseReady

	c.Client.BackingServiceInstances().Update(bsi)
	/*
		if bsi.Status.Phase != backingserviceinstanceapi.BackingServiceInstancePhaseActive {
			bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseActive

			c.Client.BackingServiceInstances().Update(bsi)
		}
	*/

	return nil
}

func checkIfPlanidExist(client osclient.Interface, planId string) (bool, *backingserviceapi.BackingService, error) {

	items, err := client.BackingServices().List(labels.Everything(), fields.Everything())
	if err != nil {
		return false, nil, err
	}

	for _, bs := range items.Items {
		for _, plans := range bs.Spec.Plans {
			if planId == plans.Id {
				fmt.Println("we found plan id at plan", bs.Spec.Name)

				return true, &bs, nil
			}
		}
	}
	return false, nil, fatalError(fmt.Sprintf("Can't find plan id %s", planId))
}
