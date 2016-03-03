package controller

import (
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
	"errors"
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
	"regexp"
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
func (c *BackingServiceInstanceController) Handle(bsi *backingserviceinstanceapi.BackingServiceInstance) (result error) {
	glog.Infoln("bsi handler called.", bsi.Name)
	
	changed := false

	switch bsi.Status.Phase {
	default:
		return errors.New("unknown phase")
	case backingserviceinstanceapi.BackingServiceInstancePhaseError:
		return errors.New("error")
	case backingserviceinstanceapi.BackingServiceInstancePhaseDestroyed:
		return nil
	case "", backingserviceinstanceapi.BackingServiceInstancePhaseCreated:
		ok, bs, err := checkIfPlanidExist(c.Client, bsi.Spec.BackingServicePlanGuid)
		if !ok {
			bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
			
			changed = true
			result =  err
			break
		//} else {
		//	bsi.Spec.BackingServiceName = bs.Name
		}
		
		bsi.Spec.BackingServiceName = bs.Spec.Name
		bsi.Spec.BackingServiceID = bs.Spec.Id
		
		for _, plan := range bs.Spec.Plans {
			if bsi.Spec.BackingServicePlanGuid == plan.Id {
				bsi.Spec.BackingServicePlanName = plan.Name
				break
			}
		}
		
		// 
		bsInstanceID := string(util.NewUUID())
		bsi.Spec.InstanceID = bsInstanceID
		bsi.Spec.Parameters = make(map[string]string)
		bsi.Spec.Parameters["instance_id"] = bsInstanceID
		
		bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseActive
		changed = true
		
	case backingserviceinstanceapi.BackingServiceInstancePhaseActive:
	
	UNBIND:
		// unbind
		if bsi.Spec.Bound && bsi.Spec.BindUuid == "" {
			servicebroker, err := servicebroker_load(c.Client, bsi.Spec.BackingServiceName)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			_, err = servicebroker_unbinding(bsi, servicebroker)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			err = deploymentconfig_clear_envs(c.Client, bsi.Spec.BindDeploymentConfig, bsi.Name)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseActive
			bsi.Spec.Bound = false
			//bsi.Spec.BindUuid = ""
			bsi.Spec.BindDeploymentConfig = ""
			bsi.Spec.Credentials = nil
			
			changed = true
		}
		
		// delete
		if bsi.Spec.InstanceID == "" {
			if bsi.Spec.Bound && bsi.Spec.BindUuid != "" {
				bsi.Spec.BindUuid = ""
				goto UNBIND
			}
			
			servicebroker, err := servicebroker_load(c.Client, bsi.Spec.BackingServiceName)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			glog.Infoln("deleting ", bsi.Name)
			if _, err = servicebroker_deprovisioning(bsi, servicebroker); err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			} else {
				err = c.Client.BackingServiceInstances(bsi.Namespace).Delete(bsi.Name)
				
				if err != nil {
					glog.Error(err)
					//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
					//
					//changed = true
					result = err
					break
				}
				
				bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseDestroyed
				changed = true
			}
		}
		
		// bind
		if bsi.Spec.Bound == false && bsi.Spec.BindUuid != "" && bsi.Spec.BindDeploymentConfig != "" {
			servicebroker, err := servicebroker_load(c.Client, bsi.Spec.BackingServiceName)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			serviceinstance := &ServiceInstance{}
			serviceinstance.ServiceId = bsi.Spec.BackingServiceID
			serviceinstance.PlanId = bsi.Spec.BackingServicePlanGuid
			serviceinstance.OrganizationGuid = bsi.Namespace
			
			svcinstance, err := servicebroker_create_instance(serviceinstance, bsi.Spec.InstanceID, servicebroker)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			} else {
				bsi.Spec.DashboardUrl = svcinstance.DashboardUrl
				glog.Infoln("create instance successfully.", svcinstance)
			}
			
			servicebinding := &ServiceBinding{
				ServiceId:      bsi.Spec.BackingServiceID,
				PlanId:         bsi.Spec.BackingServicePlanGuid,
				AppGuid:        bsi.Namespace,
				//BindResource: ,
				//Parameters: ,
				svc_instance_id: bsi.Spec.InstanceID,
			}
			
			bindingresponse, err := servicebroker_binding(servicebinding, bsi.Spec.BindUuid, servicebroker)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			err = deploymentconfig_inject_envs(c.Client, bsi.Spec.BindDeploymentConfig, bsi.Name, bindingresponse)
			if err != nil {
				glog.Error(err)
				//bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
				//
				//changed = true
				result = err
				break
			}
			
			bsi.Spec.Bound = true
			
			changed = true
		}
	}
	
	// ...

	if changed {
		c.Client.BackingServiceInstances(bsi.Namespace).Update(bsi)
	}
	
	return result
}

func servicebroker_load(c osclient.Interface, name string) (*ServiceBroker, error){
	servicebroker := &ServiceBroker{}
	if sb, err := c.ServiceBrokers().Get(name); err != nil {
		return nil, err
	} else {
		servicebroker.Url = sb.Spec.Url
		servicebroker.UserName = sb.Spec.UserName
		servicebroker.Password = sb.Spec.Password
		return servicebroker, nil
	}
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

type ServiceBroker struct {
	Url      string
	UserName string
	Password string
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

func servicebroker_create_instance(param *ServiceInstance, instance_guid string, sb *ServiceBroker) (*CreateServiceInstanceResponse, error) {
	jsonData, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(sb.UserName, sb.Password)

	resp, err := commToServiceBroker("PUT", "http://"+sb.Url+"/v2/service_instances/"+instance_guid, jsonData, header)
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

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
		if len(body) > 0 {
			err = json.Unmarshal(body, svcinstance)

			if err != nil {
				glog.Error(err)
				return nil, err
			}
		}
	}
	glog.Infof("%v,%+v\n", string(body), svcinstance)
	return svcinstance, nil
}

func servicebroker_binding(param *ServiceBinding, binding_guid string, sb *ServiceBroker) (*ServiceBindingResponse, error) {
	jsonData, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(sb.UserName, sb.Password)

	resp, err := commToServiceBroker("PUT", "http://"+sb.Url+"/v2/service_instances/"+param.svc_instance_id+"/service_bindings/"+binding_guid, jsonData, header)
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

func servicebroker_unbinding(bsi *backingserviceinstanceapi.BackingServiceInstance, sb *ServiceBroker) (interface{}, error) {

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(sb.UserName, sb.Password)

	resp, err := commToServiceBroker("DELETE", "http://"+sb.Url+"/v2/service_instances/"+bsi.Spec.InstanceID+"/service_bindings/"+bsi.Spec.BindUuid+"?service_id="+bsi.Spec.BackingServiceID+"&plan_id="+bsi.Spec.BackingServicePlanGuid, nil, header)
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

func servicebroker_deprovisioning(bsi *backingserviceinstanceapi.BackingServiceInstance, sb *ServiceBroker) (interface{}, error) {

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = basicAuthStr(sb.UserName, sb.Password)

	resp, err := commToServiceBroker("DELETE", "http://"+sb.Url+"/v2/service_instances/"+bsi.Spec.InstanceID+"?service_id="+bsi.Spec.BackingServiceID+"&plan_id="+bsi.Spec.BackingServicePlanGuid, nil, header)
	if err != nil {

		glog.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	glog.Infof("respcode from DELETE /v2/service_instances/%s: %v", bsi.Spec.InstanceID, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	type DeprovisioningResp struct {
		Response interface{}
	}
	svcDeprovisioning := &DeprovisioningResp{}

	if resp.StatusCode == http.StatusOK {
		if len(body) > 0 {
			err = json.Unmarshal(body, svcDeprovisioning)

			if err != nil {
				glog.Error(err)
				return nil, err
			}
		}
	}
	glog.Infof("%v,%+v\n", string(body), svcDeprovisioning)
	return svcDeprovisioning, nil
}

func basicAuthStr(username, password string) string {
	auth := username + ":" + password
	authstr := base64.StdEncoding.EncodeToString([]byte(auth))
	return "Basic " + authstr
}




var InvalidCharFinder = regexp.MustCompile("[^a-zA-Z0-9]")

func deploymentconfig_env_prefix(bsiName string) string {
	return fmt.Sprintf("BSI_%s_", InvalidCharFinder.ReplaceAllLiteralString(bsiName, ""))
}

func deploymentconfig_env_name(prefix string, envName string) string {
	return fmt.Sprintf("%s%s", prefix, InvalidCharFinder.ReplaceAllLiteralString(envName, "_"))
}

func deploymentconfig_inject_envs(c osclient.Interface, dcName string, bsiName string, bindingResponse *ServiceBindingResponse) error {

	
		
	return nil
}
// ENVs
//    BSI_XXXX_password=xxxxx
func deploymentconfig_clear_envs(c osclient.Interface, dcName string, bsiName string) error {
	
	return nil
}
