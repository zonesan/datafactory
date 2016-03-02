package api

import (
	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
)

const (
	// These are internal finalizer values to Origin
	FinalizerOrigin kapi.FinalizerName = "openshift.io/origin"

	// ApplicationNew is create by administrator.
	ApplicationNew ApplicationPhase = "New"

	// ApplicationRunning indicates that Application service working well.
	ApplicationActive ApplicationPhase = "Active"

	// ApplicationRunning indicates that Application service Inactive.
	ApplicationInActive ApplicationPhase = "InActive"

	// ApplicationUpdating indicates that Application is updating.
	ApplicationUpdating ApplicationPhase = "Updating"

	// ApplicationFailed indicates that Application stopped.
	ApplicationFailed ApplicationPhase = "Failed"

	// ApplicationDeletingLabel indicates that Application is delete label.
	ApplicationTerminatingLabel ApplicationPhase = "Terminating Label"

	// ApplicationFailed indicates that Application is delete all resources.
	ApplicationTerminating ApplicationPhase = "Terminating"

	// ApplicationFailed indicates that Application is delete resources error.
	ApplicationTerminatingErr ApplicationPhase = "Error"
)

var ApplicationItemSupportKinds = []string{
	"Build", "BuildConfig", "DeploymentConfig", "ImageStream", "ImageStreamTag", "ImageStreamImage", //openshift kind
	"Event", "Node", "Job", "Pod", "ReplicationController", "Service", "PersistentVolume", "PersistentVolumeClaim", //k8s kind
	"ServiceBroker", "BackingService", "BackingServiceInstance",
}

type ApplicationPhase string

type Application struct {
	unversioned.TypeMeta
	kapi.ObjectMeta

	Spec ApplicationSpec

	Status ApplicationStatus
}

type ApplicationList struct {
	unversioned.TypeMeta
	unversioned.ListMeta

	Items []Application
}

type ApplicationSpec struct {
	Name string
	//Description string
	//ImageUrl    string
	Items ItemList

	Finalizers []kapi.FinalizerName
}

type ApplicationStatus struct {
	Phase ApplicationPhase
}

type ItemList []Item

type Item struct {
	Kind   string
	Name   string
	Status string
}

const (
	ApplicationItemStatusAdd    = "add"
	ApplicationItemStatusDelete = "delete"
	ApplicationItemStatusErr    = "error"
	ApplicationItemStatusOk     = "ok"
)
