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

	ok, bs, err := checkIfPlanidExist(c.Client, bsi.Spec.BackingServicePlanGuid)
	if !ok {
		if bsi.Status.Phase != backingserviceinstanceapi.BackingServiceInstancePhaseError {
			bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
			c.Client.BackingServiceInstances().Update(bsi)
		}

		return err
	}

	sb, err := c.Client.ServiceBrokers().Get(bsi.Spec.BackingServiceName)
	if err != nil {
		return err
	}

	bsInstanceID := string(util.NewUUID())
	bsi.Spec.BackingServiceName = bs.Spec.Name
	bsi.Spec.BackingServiceID = bs.Spec.Id
	bsi.Spec.InstanceID = bsInstanceID
	bsi.Spec.Parameters = make(map[string]string)
	bsi.Spec.Parameters["instance_id"] = bsInstanceID

	sbi := &ServiceInstance{}
	sbi.ServiceId = bs.Spec.Id
	sbi.PlanId = bsi.Spec.BackingServicePlanGuid
	sbi.OrganizationGuid = bsi.Namespace

	if svcinstance, err := servicebroker_create_instance(sbi, bsInstanceID, sb.Spec.Url, sb.Spec.UserName, sb.Spec.Password); err != nil {
		glog.Errorln(err)
		bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
	} else {
		bsi.Spec.DashboardUrl = svcinstance.DashboardUrl
		bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseReady
		glog.Infoln("create instance successfully.", svcinstance)
	}

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

type ServiceInstance struct {
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

type ServiceBinding struct {
	ServiceId       string                 `json:"service_id"`
	PlanId          string                 `json:"plan_id"`
	AppGuid         string                 `json:"app_guid,omitempty"`
	BindResource    map[string]string      `json:"bind_resource,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	svc_instance_id string
}

type ServiceBindingResponse struct {
	Credentials     Credential `json:"credentials"`
	SyslogDrainUrl  string     `json:"syslog_drain_url"`
	RouteServiceUrl string     `json:"route_service_url"`
}

type Credential struct {
	Uri      string `json:"uri"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Vhost    string `json:"vhost"`
	//Database string `json:"database"`
}

func servicebroker_create_instance(param *ServiceInstance, instance_guid, broker_url, username, password string) (*CreateServiceInstanceResponse, error) {
	jsonData, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(username, password)

	resp, err := commToServiceBroker("PUT", "http://"+broker_url+"/v2/service_instances/"+instance_guid, jsonData, header)
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
				glog.Error(err)
				return nil, err
			}
		}
	}

	return svcinstance, nil
}

func servicebroker_binding(param *ServiceBinding, binding_guid, broker_url, username, password string) (*ServiceBindingResponse, error) {
	jsonData, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(username, password)

	resp, err := commToServiceBroker("PUT", "http://"+broker_url+"/v2/service_instances/"+param.svc_instance_id+"/service_bindings/"+binding_guid, jsonData, header)
	if err != nil {

		glog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	glog.Infof("respcode from PUT /v2/service_instances/%s/service_bindings/%s: %v", param.svc_instance_id, binding_guid, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	svcBinding := &ServiceBindingResponse{}

	glog.Infof("%v,%+v\n", string(body), svcBinding)
	if resp.StatusCode == http.StatusOK {
		if len(body) > 0 {
			err = json.Unmarshal(body, svcBinding)

			if err != nil {
				glog.Error(err)
				return nil, err
			}
		}
	}

	return svcBinding, nil
}

func servicebroker_unbinding(bsi *backingserviceinstanceapi.BackingServiceInstance, broker_url, username, password string) (interface{}, error) {

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(username, password)

	resp, err := commToServiceBroker("DELETE", "http://"+broker_url+"/v2/service_instances/"+bsi.Spec.InstanceID+"/service_bindings/"+bsi.Spec.BindUuid+"?service_id="+bsi.Spec.BackingServiceID+"&plan_id="+bsi.Spec.BackingServicePlanGuid, nil, header)
	if err != nil {

		glog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	glog.Infof("respcode from DELETE /v2/service_instances/%s/service_bindings/%s: %v", bsi.Spec.InstanceID, bsi.Spec.BindUuid, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	type UnBindindResp struct {
		Response interface{}
	}
	svcUnbinding := &UnBindindResp{}

	if resp.StatusCode == http.StatusOK {
		if len(body) > 0 {
			err = json.Unmarshal(body, svcUnbinding)

			if err != nil {
				glog.Error(err)
				return nil, err
			}
		}
	}
	glog.Infof("%v,%+v\n", string(body), svcUnbinding)
	return svcUnbinding, nil
}

func basicAuthStr(username, password string) string {
	auth := username + ":" + password
	authstr := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + authstr
}
