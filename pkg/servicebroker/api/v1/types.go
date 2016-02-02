package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

const (
	// These are internal finalizer values to Origin
	FinalizerOrigin kapi.FinalizerName = "openshift.io/origin"
)

type ServiceBroker struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	// Spec defines the behavior of the Namespace.
	Spec ServiceBrokerSpec `json:"spec,omitempty" description:"spec defines the behavior of the ServiceBroker"`

	// Status describes the current status of a Namespace
	Status ServiceBrokerStatus `json:"status,omitempty" description:"status describes the current status of a Project; read-only"`
}

type ServiceBrokerList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is a list of routes
	Items []ServiceBroker `json:"items" description:"list of servicebrokers"`
}

type ServiceBrokerSpec struct {
	Url      string `json:"url"`
	Name     string `json:"brokername"`
	UserName string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage
	Finalizers []kapi.FinalizerName `json:"finalizers,omitempty" description:"an opaque list of values that must be empty to permanently remove object from storage"`
}

// ProjectStatus is information about the current status of a Project
type ServiceBrokerStatus struct {
	Phase kapi.NamespacePhase `json:"phase,omitempty" description:"phase is the current lifecycle phase of the servicebroker"`
}
