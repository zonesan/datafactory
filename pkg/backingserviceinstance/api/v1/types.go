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
	Name           string   `json:"name"`
	Id             string   `json:"id"`
	Description    string   `json:"description"`
	Bindable       bool     `json:"bindable"`
	PlanUpdateable bool     `json:"plan_updateable, omitempty"`
	Tags           []string `json:"tags, omitempty"`
	Requires       []string `json:"requires, omitempty"`

	//Metadata        ServiceMetadata        `json:"metadata, omitempty"`
	Metadata        map[string]string `json:"metadata, omitempty"`
	//Plans           []ServiceInstancePlan     `json:"plans"`
	Plan            ServiceInstancePlan     `json:"plan"`
	Used            int64
	DashboardClient map[string]string `json:"dashboard_client"`
	//DashboardClient ServiceDashboardClient `json:"dashboard_client"`
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

type ServiceInstancePlan struct {
	Name        string              `json:"name"`
	Id          string              `json:"id"`
	Description string              `json:"description"`
	Metadata    ServiceInstancePlanMetadata `json:"metadata, omitempty"`
	Free        bool                `json:"free, omitempty"`
}

type ServiceInstancePlanMetadata struct {
	Bullets     []string          `json:"bullets, omitempty"`
	Costs       []ServiceInstancePlanCost `json:"costs, omitempty"`
	DisplayName string            `json:"displayName, omitempty"`
}

//TODO amount should be a array object...
type ServiceInstancePlanCost struct {
	Amount map[string]float64 `json:"amount, omitempty"`
	Unit   string             `json:"unit, omitempty"`
}

// ProjectStatus is information about the current status of a Project
type BackingServiceInstanceStatus struct {
	Phase BackingServiceInstancePhase `json:"phase,omitempty" description:"phase is the current lifecycle phase of the servicebroker"`
}

type BackingServiceInstancePhase string

const (
	BackingServiceInstancePhaseActive   BackingServiceInstancePhase = "Active"
	BackingServiceInstancePhaseInactive BackingServiceInstancePhase = "Inactive"
)
