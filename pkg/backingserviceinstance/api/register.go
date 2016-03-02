package api

import (
	"k8s.io/kubernetes/pkg/api"
)

func init() {
	api.Scheme.AddKnownTypes("",
		&BackingServiceInstance{},
		&BackingServiceInstanceList{},
		&BindingRequest{},
	)
}

func (*BackingServiceInstance) IsAnAPIObject()     {}
func (*BackingServiceInstanceList) IsAnAPIObject() {}
func (*BindingRequest) IsAnAPIObject()             {}