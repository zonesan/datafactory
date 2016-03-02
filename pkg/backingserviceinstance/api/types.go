package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type BackingServiceInstance struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Spec defines the behavior of the Namespace.
	Spec BackingServiceInstanceSpec

	// Status describes the current status of a Namespace
	Status BackingServiceInstanceStatus
}

type BackingServiceInstanceList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	// Items is a list of routes
	Items []BackingServiceInstance
}

/*
type BackingServiceInstanceSpec struct {
	Config                 map[string]string `json:"config, omitempty"`
	DashboardUrl           string            `json:"dashboard_url, omitempty"`
	BackingServiceGuid     string            `json:"backingservice_guid, omitempty"`
	BackingServicePlanGuid string            `json:"backingservice_plan_guid, omitempty"`
	Parameters             map[string]string `json:"parameters, omitempty"`
	Binding                bool              `json:"binding, omitempty"`
	BindUuid               string            `json:"bind_uuid, omitempty"`
	BindDeploymentConfig   map[string]string `json:"bind_deploymentconfig, omitempty"`
	Credential             map[string]string `json:"credential, omitempty"`
	Tags                   []string          `json:"tags, omitempty"`
}
*/

type BackingServiceInstanceSpec struct {
	InstanceProvisioning
	InstanceBinding
	Bound      bool
	InstanceID string
	Tags       []string
}

type InstanceProvisioning struct {
	DashboardUrl           string
	BackingServiceName     string
	BackingServiceID       string
	BackingServicePlanGuid string
	BackingServicePlanName string
	Parameters             map[string]string
}

type InstanceBinding struct {
	BindUuid             string
	BindDeploymentConfig string
	Credentials          map[string]string
}

// ProjectStatus is information about the current status of a Project
type BackingServiceInstanceStatus struct {
	Phase BackingServiceInstancePhase
}

type BackingServiceInstancePhase string

const (
	BackingServiceInstancePhaseActive   BackingServiceInstancePhase = "Active"
	BackingServiceInstancePhaseCreated  BackingServiceInstancePhase = "Created"
	BackingServiceInstancePhaseInactive BackingServiceInstancePhase = "Inactive"
	BackingServiceInstancePhaseModified BackingServiceInstancePhase = "Modified"
	BackingServiceInstancePhaseReady    BackingServiceInstancePhase = "Ready"
	BackingServiceInstancePhaseError    BackingServiceInstancePhase = "Error"
)

//=====================================================
// 
//=====================================================

const BindKind_DeploymentConfig = "DeploymentConfig"

//type BindingRequest struct {
//	unversioned.TypeMeta
//	kapi.ObjectMeta
//
//	// the dc
//	DeploymentConfigName string `json:"deployment_name, omitempty"`
//}

type BindingRequestOptions struct {
	unversioned.TypeMeta
	kapi.ObjectMeta
	
	
	
	BindKind            string `json:"bindKind, omitempty"`
	BindResourceVersion string `json:"bindResourceVersion, omitempty"`
	ResourceName        string `json:"resourceName, omitempty"`
}

func NewBindingRequestOptions (kind, version, name string) *BindingRequestOptions {
	return &BindingRequestOptions{
		BindKind: kind,
		BindResourceVersion: version,
		ResourceName: name,
	}
}
