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
	InstanceID             string            `json:"instance_id, omitempty"`
	DashboardUrl           string            `json:"dashboard_url, omitempty"`
	BackingServiceName     string            `json:"backingservice_name, omitempty"`
	BackingServiceID       string            `json:"backingservice_id, omitempty"`
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
	InstanceProvisioning `json:"provisioning, omitempty"`
	InstanceBinding      `json:"binding, omitempty"`
	Bound                bool     `json:"bound, omitempty"`
	InstanceID           string   `json:"instance_id, omitempty"`
	Tags                 []string `json:"tags, omitempty"`
}

type InstanceProvisioning struct {
	DashboardUrl           string            `json:"dashboard_url, omitempty"`
	BackingServiceName     string            `json:"backingservice_name, omitempty"`
	BackingServiceID       string            `json:"backingservice_id, omitempty"`
	BackingServicePlanGuid string            `json:"backingservice_plan_guid, omitempty"`
	BackingServicePlanName string            `json:"backingservice_plan_name, omitempty"`
	Parameters             map[string]string `json:"parameters, omitempty"`
}

type InstanceBinding struct {
	BindUuid             string            `json:"bind_uuid, omitempty"`
	BindDeploymentConfig string            `json:"bind_deploymentconfig, omitempty"`
	Credentials          map[string]string `json:"credential, omitempty"`
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

type BindingRequest struct {
	unversioned.TypeMeta
	// TODO: build request should allow name generation via Name and GenerateName, build config
	// name should be provided as a separate field
	kapi.ObjectMeta

	// the application to be bound
	//app *Application
	ApplicationUuid string `json:"application_uuid, omitempty"`
}
