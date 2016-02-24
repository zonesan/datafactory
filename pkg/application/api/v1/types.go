package v1

import (
	"k8s.io/kubernetes/pkg/api/unversioned"
	kapi "k8s.io/kubernetes/pkg/api/v1"
)

const (
	// These are internal finalizer values to Origin
	FinalizerOrigin kapi.FinalizerName = "openshift.io/origin"

	// ApplicationNew is create by administrator.
	ApplicationNew ApplicationPhase = "New"

	// ApplicationRunning indicates that Application service working well.
	ApplicationActive ApplicationPhase = "Active"

	// ApplicationFailed indicates that Application stopped.
	ApplicationFailed ApplicationPhase = "Failed"
)

var ApplicationItemSupportKinds = []string{
	"Build", "BuildConfig", "DeploymentConfig", "ImageStream", "ImageStreamTag", "ImageStreamImage", //openshift kind
	"Event", "Node", "Pod", "ReplicationController", "Service", "PersistentVolume", "PersistentVolumeClaim", //k8s kind
	"ServiceBroker", "BackingService", "BackingServiceInstance",
}

type ApplicationPhase string

type Application struct {
	unversioned.TypeMeta `json:",inline"`
	kapi.ObjectMeta      `json:"metadata,omitempty"`

	// Spec defines the behavior of the Namespace.
	Spec ApplicationSpec `json:"spec,omitempty" description:"spec defines the behavior of the Application"`

	// Status describes the current status of a Namespace
	Status ApplicationStatus `json:"status,omitempty" description:"status describes the current status of a Application; read-only"`
}

type ApplicationList struct {
	unversioned.TypeMeta `json:",inline"`
	unversioned.ListMeta `json:"metadata,omitempty"`

	// Items is a list of applications
	Items []Application `json:"items" description:"list of Applications"`
}

type ApplicationSpec struct {
	Name        string   `json:"name" description:"name defines the name of a Application"`
	Description string   `json:"description" description:"description defines the description of a Application"`
	ImageUrl    string   `json:"imageUrl" description:"imageUrl defines the image url of a Application"`
	Items       ItemList `json:"items" description:"items defines the resources to be labeled in a Application"`
	// Finalizers is an opaque list of values that must be empty to permanently remove object from storage
	Finalizers []kapi.FinalizerName `json:"finalizers,omitempty" description:"an opaque list of values that must be empty to permanently remove object from storage"`
}

// ApplicationStatus is information about the current status of a Application
type ApplicationStatus struct {
	Phase ApplicationPhase `json:"phase,omitempty" description:"phase is the current lifecycle phase of the Application"`
}

type ItemList []Item

type Item struct {
	Kind string `json:"kind" description:"kind defines the item kind of a item in Application"`
	Name string `json:"name" description:"name defines the item name of a item in Application"`
}
