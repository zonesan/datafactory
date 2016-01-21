package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

type BackingService struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	// Spec defines the behavior of the Namespace.
	Spec BackingServiceSpec

	// Status describes the current status of a Namespace
	Status BackingServiceStatus
}

type BackingServiceList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	// Items is a list of routes
	Items []BackingService
}

type BackingServiceSpec struct {
	Url      string
	Name     string
	UserName string
	Password string
}

// BackingServiceStatus is information about the current status of a Project
type BackingServiceStatus struct {
	Phase BackingServicePhase
}

type BackingServicePhase string
