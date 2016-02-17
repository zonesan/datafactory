package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	// These are internal finalizer values to Origin
	FinalizerOrigin kapi.FinalizerName = "openshift.io/origin"
)

type ServiceBroker struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Spec defines the behavior of the Namespace.
	Spec ServiceBrokerSpec

	// Status describes the current status of a Namespace
	Status ServiceBrokerStatus
}

type ServiceBrokerList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	Items []ServiceBroker
}

type ServiceBrokerSpec struct {
	Url      string
	Name     string
	UserName string
	Password string

	Finalizers []kapi.FinalizerName
}

// ProjectStatus is information about the current status of a Project
type ServiceBrokerStatus struct {
	Phase kapi.NamespacePhase
}
