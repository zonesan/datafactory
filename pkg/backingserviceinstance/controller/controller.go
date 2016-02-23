package controller

import (
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	backingserviceinstanceapi "github.com/openshift/origin/pkg/backingserviceinstance/api"
	osclient "github.com/openshift/origin/pkg/client"
	"io/ioutil"
	kclient "k8s.io/kubernetes/pkg/client/unversioned"
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
	glog.Infoln("bsi handler called.", bsi.Name)
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

	sb, err := c.Client.ServiceBrokers().Get(bsi.Spec.BackingServiceName)
	if err != nil {
		return err
	}

	bsInstanceID := string(util.NewUUID())
	bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseReady
	bsi.Spec.Parameters = make(map[string]string)
	bsi.Spec.Parameters["instance_id"] = bsInstanceID

	sbi := &SBServiceInstance{}
	sbi.ServiceId = bsInstanceID
	sbi.PlanId = bsi.Spec.BackingServicePlanGuid
	sbi.OrganizationGuid = bsi.Namespace

	if svcinstance, err := servicebroker_create_instance(sbi, bsInstanceID, sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password); err != nil {
		glog.Errorln(err)

	} else {

		glog.Infoln("create instance successfully.", svcinstance)
	}

	/*
		resp, err := commToServiceBroker("PUT", sbInstanceUrl, jsonData, nil)
		if err != nil {
			//glog.Error(err)
			fmt.Println(err)
			return err
		}
		defer resp.Body.Close()
	*/

	c.Client.BackingServiceInstances().Update(bsi)
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
				glog.Info("we found plan id at plan", bs.Spec.Name)

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

type SBServiceInstance struct {
	ServiceId        string `json:"service_id"`
	PlanId           string `json:"plan_id"`
	OrganizationGuid string `json:"organization_guid"`
	SpaceGuid        string `json:"space_guid"`
	//Incomplete       bool        `json:"accepts_incomplete, omitempty"`
	Parameters interface{} `json:"parameters, omitempty"`
}
type LastOperation struct {
	State                    string `json:"state"`
	Description              string `json:"description"`
	AsyncPollIntervalSeconds int    `json:"async_poll_interval_seconds, omitempty"`
}
type CreateServiceInstanceResponse struct {
	DashboardUrl  string         `json:"dashboard_url"`
	LastOperation *LastOperation `json:"last_operation, omitempty"`
}

func servicebroker_create_instance(param *SBServiceInstance, instance_guid, broker_url, username, password string) (*CreateServiceInstanceResponse, error) {
	jsonData, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(username, password)

	resp, err := commToServiceBroker("PUT", broker_url+"/v2/service_instances/"+instance_guid, jsonData, header)
	if err != nil {

		glog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	glog.Infof("respcode from /v2/service_instances/%s: %v", instance_guid, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	svcinstance := &CreateServiceInstanceResponse{}

	glog.Infof("%v,%+v\n", string(body), svcinstance)
	if resp.StatusCode == http.StatusOK {
		if len(body) > 0 {
			err = json.Unmarshal(body, svcinstance)

			if err != nil {
				fmt.Println(err)
				return nil, err
			}
		}
	}

	glog.Info("%v,%+v\n", string(body), svcinstance)

	return svcinstance, nil
}

func basicAuthStr(username, password string) string {
	auth := username + ":" + password
	authstr := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + authstr
}
