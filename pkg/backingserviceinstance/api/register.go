package api

import (
	"k8s.io/kubernetes/pkg/api"
)

func init() {
	api.Scheme.AddKnownTypes("",
		&BackingServiceInstance{},
		&BackingServiceInstanceList{},
	)
}

func (*BackingServiceInstance) IsAnAPIObject()     {}
func (*BackingServiceInstanceList) IsAnAPIObject() {}
