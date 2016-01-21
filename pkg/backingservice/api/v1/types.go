package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

type BackingService struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	// Spec defines the behavior of the Namespace.
	Spec BackingServiceSpec `json:"spec,omitempty" description:"spec defines the behavior of the ServiceBroker"`

	// Status describes the current status of a Namespace
	Status BackingServiceStatus `json:"status,omitempty" description:"status describes the current status of a Project; read-only"`
}

type BackingServiceList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is a list of routes
	Items []BackingService `json:"items" description:"list of servicebrokers"`
}

type BackingServiceSpec struct {
	Url      string `json:"url"`
	Name     string `json:"brokername"`
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ProjectStatus is information about the current status of a Project
type BackingServiceStatus struct {
	Phase BackingServicePhase `json:"phase,omitempty" description:"phase is the current lifecycle phase of the servicebroker"`
}
type BackingServicePhase string
