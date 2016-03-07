package controller

import (
	backingserviceapi "github.com/openshift/origin/pkg/backingservice/api"
	//"errors"
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
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api"
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
	
	binduuid := ""
	
	// init
	if len(bsi.Spec.Parameters) == 0 {
		glog.Infoln("to init bsi ", bsi.Name)
		
		bsi.Spec.Parameters = make(map[string]string)
		
		//ok, bs, err := checkIfPlanidExist(c.Client, bsi.Spec.BackingServicePlanGuid)
		//if !ok {
		//	result = err
		//	goto ERROR
		////} else {
		////	bsi.Spec.BackingServiceName = bs.Name
		//}
		
		bs, err := c.Client.BackingServices().Get(bsi.Spec.BackingServiceName)
		if err != nil {
			result = err
			goto ERROR
		}
		
		//bsi.Spec.BackingServiceName = bs.Spec.Name
		//bsi.Spec.BackingServiceID = bs.Spec.Id
		
		plan_found := false
		for _, plan := range bs.Spec.Plans {
			if bsi.Spec.BackingServicePlanGuid == plan.Id {
				//bsi.Spec.BackingServicePlanName = plan.Name
				plan_found = true
				break
			}
		}
		
		if ! plan_found {
			result = fmt.Errorf("plan (%s) in bs(%s) for bsi (%s) not found", bsi.Spec.BackingServicePlanGuid, bsi.Spec.BackingServiceName, bsi.Name)
			goto ERROR
		}
		
		glog.Infoln("bsi inited. ", bsi.Name)
		
		// 
		bsInstanceID := string(util.NewUUID())
		bsi.Spec.InstanceID = bsInstanceID
		bsi.Spec.Parameters["instance_id"] = bsInstanceID
		
		bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseCreated
		changed = true
	}
	
	binduuid = bsi.Spec.BindUuid // avoid failed to unbind but etc saved
	
UNBIND:

	// unbind
	if bsi.Spec.Bound && binduuid == "" {
		glog.Infoln("bsi to unbind ", bsi.Name)
		
		servicebroker, err := servicebroker_load(c.Client, bsi.Spec.BackingServiceName)
		if err != nil {
			result = err
			goto ERROR
		}
		
		glog.Infoln("servicebroker_unbinding")
		
		_, err = servicebroker_unbinding(bsi, servicebroker)
		if err != nil {
			result = err
			goto ERROR
		}
		
		glog.Infoln("deploymentconfig_clear_envs")
		
		err = c.deploymentconfig_clear_envs(bsi)
		if err != nil {
			result = err
			goto ERROR
		}
		
		glog.Infoln("bsi is unbound ", bsi.Name)
		
		bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseActive
		bsi.Spec.Bound = false
		bsi.Spec.BoundTime = nil
		//bsi.Spec.BindUuid = ""
		binduuid = ""
		bsi.Spec.BindDeploymentConfig = ""
		bsi.Spec.Credentials = nil
		
		changed = true
	}
	
	// delete
	if bsi.Spec.InstanceID == "" {
		glog.Infoln("bsi to delete ", bsi.Name)
		
		if bsi.Spec.Bound && binduuid != "" {
			//bsi.Spec.InstanceID = "" // bug
			binduuid = ""
			goto UNBIND
		}
		
		if bsi.Spec.BackingServiceName != "" {
			servicebroker, err := servicebroker_load(c.Client, bsi.Spec.BackingServiceName)
			if err != nil {
				result = err
				goto ERROR
			}
			
			glog.Infoln("deleting ", bsi.Name)
			if _, err = servicebroker_deprovisioning(bsi, servicebroker); err != nil {
				result = err
				goto ERROR
			}
		}
		
		glog.Infoln("bsi deleted ", bsi.Name)
		
		bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseInactive
		changed = true
	}
	
	// delete etcd
	if bsi.Status.Phase == backingserviceinstanceapi.BackingServiceInstancePhaseInactive {
		err := c.Client.BackingServiceInstances(bsi.Namespace).Delete(bsi.Name)
		
		// Status.Phase should not change to BackingServiceInstancePhaseError
		//if err != nil {
		//	result = err
		//	goto ERROR
		//}
		
		glog.Infoln("bsi delete etcd ", bsi.Name, ", error: ", err)
		
		result = err
		goto END
	}
	
	// bind
	if bsi.Spec.Bound == false && bsi.Spec.BindUuid != "" && bsi.Spec.BindDeploymentConfig != "" {
		glog.Infoln("bsi to bind ", bsi.Name, " and ", bsi.Spec.BindDeploymentConfig)
		
		servicebroker, err := servicebroker_load(c.Client, bsi.Spec.BackingServiceName)
		if err != nil {
			result = err
			goto ERROR
		}
		
		serviceinstance := &ServiceInstance{}
		serviceinstance.ServiceId = bsi.Spec.BackingServiceSpecID
		serviceinstance.PlanId = bsi.Spec.BackingServicePlanGuid
		serviceinstance.OrganizationGuid = bsi.Namespace
		
		glog.Infoln("servicebroker_create_instance")
		
		svcinstance, err := servicebroker_create_instance(serviceinstance, bsi.Spec.InstanceID, servicebroker)
		if err != nil {
			result = err
			goto ERROR
		} else {
			bsi.Spec.DashboardUrl = svcinstance.DashboardUrl
			glog.Infoln("create instance successfully.", svcinstance)
		}
		
		servicebinding := &ServiceBinding{
			ServiceId:      bsi.Spec.BackingServiceSpecID,
			PlanId:         bsi.Spec.BackingServicePlanGuid,
			AppGuid:        bsi.Namespace,
			//BindResource: , 
			//Parameters: , 
			svc_instance_id: bsi.Spec.InstanceID,
		}
		
		if len(bsi.Spec.Credentials) == 0 {
			glog.Infoln("servicebroker_binding")
			
			bindingresponse, err := servicebroker_binding(servicebinding, bsi.Spec.BindUuid, servicebroker)
			if err != nil {
				result = err
				goto ERROR
			}
			
			bsi.Spec.Credentials = make(map[string]string)
			bsi.Spec.Credentials["Uri"] = bindingresponse.Credentials.Uri
			bsi.Spec.Credentials["Name"] = bindingresponse.Credentials.Name
			bsi.Spec.Credentials["Username"] = bindingresponse.Credentials.Username
			bsi.Spec.Credentials["Password"] = bindingresponse.Credentials.Password
			bsi.Spec.Credentials["Host"] = bindingresponse.Credentials.Host
			bsi.Spec.Credentials["Port"] = bindingresponse.Credentials.Port
			bsi.Spec.Credentials["Vhost"] = bindingresponse.Credentials.Vhost
			// = bindingresponse.SyslogDrainUrl
			// = bindingresponse.RouteServiceUrl
			
			changed = true
		}
		
		glog.Infoln("deploymentconfig_inject_envs")
		
		err = c.deploymentconfig_inject_envs(bsi)
		if err != nil {
			result = err
			goto ERROR
		}
		
		glog.Infoln("bsi bound. ", bsi.Name)
		
		bsi.Spec.Bound = true
		now := unversioned.Now()
		bsi.Spec.BoundTime = &now //&unversioned.Now()
		
		changed = true
	}

	// ...
	
	goto END
	
ERROR:
	
	glog.Error(result)
	changed = changed || bsi.Status.Phase != backingserviceinstanceapi.BackingServiceInstancePhaseError
	bsi.Status.Phase = backingserviceinstanceapi.BackingServiceInstancePhaseError
	
END:

	if changed {
		glog.Infoln("bsi etc changed and update. ")
		
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

	resp, err := commToServiceBroker("DELETE", "http://"+sb.Url+"/v2/service_instances/"+bsi.Spec.InstanceID+"/service_bindings/"+bsi.Spec.BindUuid+"?service_id="+bsi.Spec.BackingServiceSpecID+"&plan_id="+bsi.Spec.BackingServicePlanGuid, nil, header)
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

	resp, err := commToServiceBroker("DELETE", "http://"+sb.Url+"/v2/service_instances/"+bsi.Spec.InstanceID+"?service_id="+bsi.Spec.BackingServiceSpecID+"&plan_id="+bsi.Spec.BackingServicePlanGuid, nil, header)
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

func (c *BackingServiceInstanceController) deploymentconfig_inject_envs(bsi *backingserviceinstanceapi.BackingServiceInstance) error {
	return c.deploymentconfig_modify_envs(bsi, true)
}

func (c *BackingServiceInstanceController) deploymentconfig_clear_envs(bsi *backingserviceinstanceapi.BackingServiceInstance) error {
	return c.deploymentconfig_modify_envs(bsi, false)
}

// return overritten or not
func env_set(envs []kapi.EnvVar, envName, envValue string) (bool, []kapi.EnvVar) {
	if envs == nil {
		envs = []kapi.EnvVar{}
	}
	
	for i := len(envs) - 1; i >= 0; i -- {
		if envs[i].Name == envName {
			envs[i] = kapi.EnvVar{Name: envName, Value: envValue}
			return true, envs
		}
	}
	
	envs = append(envs, kapi.EnvVar{Name: envName, Value: envValue})
	return false, envs
}

// return unset or not
func env_unset(envs []kapi.EnvVar, envName string) (bool, []kapi.EnvVar) {
	if envs == nil {
		return false, nil
	}
	
	n := len(envs)
	index := 0
	for i := 0; i < n; i ++ {
		if envs[i].Name != envName {
			if index < i {
				envs[index] = envs[i]
			}
			index ++
		}
	}
	
	return index < n, envs[:index]
}

func (c *BackingServiceInstanceController) deploymentconfig_modify_envs(bsi *backingserviceinstanceapi.BackingServiceInstance, toInject bool) error {
	dc, err := c.Client.DeploymentConfigs(bsi.Namespace).Get(bsi.Spec.BindDeploymentConfig)
	if err != nil {
		return err
	}
	
	if dc.Spec.Template == nil {
		return nil
	}
	//pod_tempalte, err := c.KubeClient.PodTemplates(bsi.Namespace).Get(dc.Spec.Template.Name)
	//if err != nil {
	//	return err
	//}
	
	env_prefix := deploymentconfig_env_prefix(bsi.Name)
	containers := dc.Spec.Template.Spec.Containers
	
	if toInject {
		for _, c := range containers {
			for k, v := range bsi.Spec.Credentials {
				_, c.Env = env_set(c.Env, deploymentconfig_env_name(env_prefix, k), v)
			}
		}
	} else {
		for _, c := range containers {
			for k, _ := range bsi.Spec.Credentials {
				_, c.Env = env_unset(c.Env, deploymentconfig_env_name(env_prefix, k))
			}
		}
	}

	//if _, err := c.KubeClient.PodTemplates(bsi.Namespace).Update(pod_tempalte); err != nil {
	//	return err
	//}
	if _, err := c.Client.DeploymentConfigs(bsi.Namespace).Update(dc); err != nil {
		return err
	}
	
	c.deploymentconfig_print_envs(bsi)
	
	return nil
}

func (c *BackingServiceInstanceController) deploymentconfig_print_envs(bsi *backingserviceinstanceapi.BackingServiceInstance) {
	dc, err := c.Client.DeploymentConfigs(bsi.Namespace).Get(bsi.Spec.BindDeploymentConfig)
	if err != nil {
		fmt.Println("dc not found: ", bsi.Spec.BindDeploymentConfig)
		return
	}
	
	if dc.Spec.Template == nil {
		fmt.Println("dc.Spec.Template is nil")
		return
	}
	
	containers := dc.Spec.Template.Spec.Containers
	
	for _, c := range containers {
		fmt.Println("**********  envs in container")
		
		for _, env := range c.Env {
			fmt.Println("     env[", env.Name, ",] = ", env.Value)
		}
	}
}