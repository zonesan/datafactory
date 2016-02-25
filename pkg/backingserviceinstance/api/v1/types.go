package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

type BackingServiceInstance struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	// Spec defines the behavior of the Namespace.
	Spec BackingServiceInstanceSpec `json:"spec,omitempty" description:"spec defines the behavior of the BackingServiceInstance"`

	// Status describes the current status of a Namespace
	Status BackingServiceInstanceStatus `json:"status,omitempty" description:"status describes the current status of a Project; read-only"`
}

type BackingServiceInstanceList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is a list of routes
	Items []BackingServiceInstance `json:"items" description:"list of BackingServiceInstances"`
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
	Provisioning InstanceProvisioning `json:"provisioning, omitempty"`
	Binding      InstanceBinding      `json:"binding, omitempty"`
}

type InstanceProvisioning struct {
	DashboardUrl           string            `json:"dashboard_url, omitempty"`
	BackingServiceName     string            `json:"backingservice_name, omitempty"`
	BackingServicePlanGuid string            `json:"backingservice_plan_guid, omitempty"`
	Parameters             map[string]string `json:"parameters, omitempty"`
}

type InstanceBinding struct {
	BindUuid                     string            `json:"bind_uuid, omitempty"`
	InstanceBindDeploymentConfig map[string]string `json:"bind_deploymentconfig, omitempty"`
	Credential                   map[string]string `json:"credential, omitempty"`
}

type InstanceBindDeploymentConfig struct {
	Parameters map[string]string `json:"parameters, omitempty"`
}

type BackingServiceInstanceStatus struct {
	Phase BackingServiceInstancePhase `json:"phase, omitempty"`
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
