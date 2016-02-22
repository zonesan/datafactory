package controller

import (
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"

	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	osclient "github.com/openshift/origin/pkg/client"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"

	"bytes"
	"encoding/json"
	"fmt"
	"k8s.io/kubernetes/pkg/fields"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/util"
	"net/http"
	"strings"
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
		bsi.Spec.BackingServiceName = bs.ObjectMeta.Name
	}

	bsInstanceID := string(util.NewUUID())
	bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseReady
	bsi.Spec.Parameters = make(map[string]string)
	bsi.Spec.Parameters["instance_id"] = bsInstanceID

	type sbsvcinstance struct {
		ServiceId        string `json:"service_id"`
		PlanId           string `json:"plan_id"`
		OrganizationGuid string `json:"organization_guid"`
		SpaceGuid        string `json:"space_guid"`
		//Incomplete       bool        `json:"accepts_incomplete, omitempty"`
		Parameters interface{} `json:"parameters, omitempty"`
	}

	sbInstanceUrl := "http://$ServiceBrokerUrl" + "/v2/service_instances/" + bsInstanceID
	fmt.Println(sbInstanceUrl)

	jsonData, err := json.Marshal(sbsvcinstance{})
	if err != nil {
		return err
	}
	resp, err := commToServiceBroker("PUT", sbInstanceUrl, jsonData, nil)
	if err != nil {
		//glog.Error(err)
		fmt.Println(err)
		return err
	}
	defer resp.Body.Close()


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


func commToServiceBroker(method, path string, jsonData []byte, header map[string]string) (resp *http.Response, err error) {

	fmt.Println(method, path, string(jsonData))

	req, err := http.NewRequest(strings.ToUpper(method) /*SERVICE_BROKER_API_SERVER+*/, path, bytes.NewBuffer(jsonData))

	if len(header) > 0 {
		for key, value := range header {
			req.Header.Set(key, value)
		}
	}

	return http.DefaultClient.Do(req)
}

