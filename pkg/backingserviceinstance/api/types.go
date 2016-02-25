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

	// Items is a list of BackingServiceInstance
	Items []BackingServiceInstance
}

/*
type BackingServiceInstanceSpec struct {
	Config                 map[string]string
	InstanceID             string
	DashboardUrl           string
	BackingServiceName     string
	BackingServiceID       string
	BackingServicePlanGuid string
	Parameters             map[string]string
	Binding                bool
	BindUuid               string
	BindDeploymentConfig   map[string]string
	Credential             map[string]string
	Tags                   []string
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
	Parameters             map[string]string
}

type InstanceBinding struct {
	BindUuid                     string
	InstanceBindDeploymentConfig map[string]string
	Credential                   map[string]string
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
