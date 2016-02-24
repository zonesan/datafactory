package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

type BackingServiceInstance struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	// Spec defines the behavior of the Namespace.
	Spec BackingServiceInstanceSpec `json:"spec,omitempty" description:"spec defines the behavior of the ServiceBroker"`

	// Status describes the current status of a Namespace
	Status BackingServiceInstanceStatus `json:"status,omitempty" description:"status describes the current status of a Project; read-only"`
}

type BackingServiceInstanceList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is a list of routes
	Items []BackingServiceInstance `json:"items" description:"list of servicebrokers"`
}

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

//type ServiceMetadata struct {
//	DisplayName         string `json:"displayName, omitempty"`
//	ImageUrl            string `json:"imageUrl, omitempty"`
//	LongDescription     string `json:"longDescription, omitempty"`
//	ProviderDisplayName string `json:"providerDisplayName, omitempty"`
//	DocumentationUrl    string `json:"documentationUrl, omitempty"`
//	SupportUrl          string `json:"supportUrl, omitempty"`
//}

//type ServiceDashboardClient struct {
//	Id          string `json:"id, omitempty"`
//	Secret      string `json:"secret, omitempty"`
//	RedirectUri string `json:"redirect_uri, omitempty"`
//}

// ProjectStatus is information about the current status of a Project
type BackingServiceInstanceStatus struct {
	Phase BackingServiceInstancePhase `json:"phase,omitempty" description:"phase is the current lifecycle phase of the servicebroker"`
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
